package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/bipulkrdas/orgmind/backend/internal/config"
	"github.com/bipulkrdas/orgmind/backend/internal/models"
	"github.com/bipulkrdas/orgmind/backend/internal/repository"
	"github.com/bipulkrdas/orgmind/backend/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// ErrInvalidCredentials is returned when email or password is incorrect
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserAlreadyExists is returned when attempting to create a user with an existing email
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	// ErrUnsupportedProvider is returned when an unsupported OAuth provider is specified
	ErrUnsupportedProvider = errors.New("unsupported OAuth provider")
	// ErrOAuthFailed is returned when OAuth authentication fails
	ErrOAuthFailed = errors.New("OAuth authentication failed")
)

// authService implements AuthService interface
type authService struct {
	userRepo       repository.UserRepository
	resetTokenRepo repository.PasswordResetTokenRepository
	cfg            *config.Config
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(userRepo repository.UserRepository, resetTokenRepo repository.PasswordResetTokenRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo:       userRepo,
		resetTokenRepo: resetTokenRepo,
		cfg:            cfg,
	}
}

// SignUp creates a new user account with email and password
func (s *authService) SignUp(ctx context.Context, email, password, firstName, lastName string) (*models.User, string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, "", ErrUserAlreadyExists
	}

	// Hash password with bcrypt cost 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	hashedPasswordStr := string(hashedPassword)
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: &hashedPasswordStr,
		FirstName:    &firstName,
		LastName:     &lastName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user to database
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		// Check if error is due to duplicate email
		if err.Error() == fmt.Sprintf("user with email %s already exists", email) {
			return nil, "", ErrUserAlreadyExists
		}
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTExpirationHours)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

// SignIn authenticates a user with email and password
func (s *authService) SignIn(ctx context.Context, email, password string) (string, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// Check if user has a password (not OAuth-only user)
	if user.PasswordHash == nil {
		return "", ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTExpirationHours)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// InitiateOAuth initiates OAuth flow for the specified provider
func (s *authService) InitiateOAuth(ctx context.Context, provider string) (string, error) {
	oauthConfig, err := s.getOAuthConfig(provider)
	if err != nil {
		return "", err
	}

	// Generate state token for CSRF protection
	state := uuid.New().String()

	// Generate authorization URL
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return authURL, nil
}

// HandleOAuthCallback handles OAuth callback and creates/authenticates user
func (s *authService) HandleOAuthCallback(ctx context.Context, provider, code string) (string, error) {
	oauthConfig, err := s.getOAuthConfig(provider)
	if err != nil {
		return "", err
	}

	// Exchange authorization code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("%w: failed to exchange code: %v", ErrOAuthFailed, err)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfoFromProvider(ctx, provider, token)
	if err != nil {
		return "", fmt.Errorf("%w: failed to get user info: %v", ErrOAuthFailed, err)
	}

	// Check if user already exists
	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		// User doesn't exist, create new user
		providerName := provider
		user = &models.User{
			ID:            uuid.New().String(),
			Email:         userInfo.Email,
			FirstName:     &userInfo.FirstName,
			LastName:      &userInfo.LastName,
			OAuthProvider: &providerName,
			OAuthID:       &userInfo.ID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists, update OAuth info if needed
		if user.OAuthProvider == nil || user.OAuthID == nil {
			providerName := provider
			user.OAuthProvider = &providerName
			user.OAuthID = &userInfo.ID
			user.UpdatedAt = time.Now()

			err = s.userRepo.Update(ctx, user)
			if err != nil {
				return "", fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateToken(user.ID, user.Email, s.cfg.JWTSecret, s.cfg.JWTExpirationHours)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return jwtToken, nil
}

// getOAuthConfig returns OAuth2 config for the specified provider
func (s *authService) getOAuthConfig(provider string) (*oauth2.Config, error) {
	switch provider {
	case "google":
		return &oauth2.Config{
			ClientID:     s.cfg.GoogleClientID,
			ClientSecret: s.cfg.GoogleClientSecret,
			RedirectURL:  s.cfg.OAuthRedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		}, nil

	case "okta":
		return &oauth2.Config{
			ClientID:     s.cfg.OktaClientID,
			ClientSecret: s.cfg.OktaClientSecret,
			RedirectURL:  s.cfg.OAuthRedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("https://%s/oauth2/v1/authorize", s.cfg.OktaDomain),
				TokenURL: fmt.Sprintf("https://%s/oauth2/v1/token", s.cfg.OktaDomain),
			},
		}, nil

	case "office365":
		return &oauth2.Config{
			ClientID:     s.cfg.Office365ClientID,
			ClientSecret: s.cfg.Office365ClientSecret,
			RedirectURL:  s.cfg.OAuthRedirectURL,
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
		}, nil

	default:
		return nil, ErrUnsupportedProvider
	}
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

// getUserInfoFromProvider fetches user information from OAuth provider
func (s *authService) getUserInfoFromProvider(ctx context.Context, provider string, token *oauth2.Token) (*OAuthUserInfo, error) {
	switch provider {
	case "google":
		return s.getGoogleUserInfo(ctx, token)
	case "okta":
		return s.getOktaUserInfo(ctx, token)
	case "office365":
		return s.getOffice365UserInfo(ctx, token)
	default:
		return nil, ErrUnsupportedProvider
	}
}

// getGoogleUserInfo fetches user info from Google
func (s *authService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		GivenName string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.ID,
		Email:     data.Email,
		FirstName: data.GivenName,
		LastName:  data.FamilyName,
	}, nil
}

// getOktaUserInfo fetches user info from Okta
func (s *authService) getOktaUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get(fmt.Sprintf("https://%s/oauth2/v1/userinfo", s.cfg.OktaDomain))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Sub       string `json:"sub"`
		Email     string `json:"email"`
		GivenName string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.Sub,
		Email:     data.Email,
		FirstName: data.GivenName,
		LastName:  data.FamilyName,
	}, nil
}

// getOffice365UserInfo fetches user info from Office365
func (s *authService) getOffice365UserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        string `json:"id"`
		Mail      string `json:"mail"`
		GivenName string `json:"givenName"`
		Surname   string `json:"surname"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.ID,
		Email:     data.Mail,
		FirstName: data.GivenName,
		LastName:  data.Surname,
	}, nil
}

// ResetPassword generates a reset token and sends password reset email
func (s *authService) ResetPassword(ctx context.Context, email string) error {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// Generate secure random token
	tokenStr := uuid.New().String()

	// Create password reset token with 1 hour expiration
	resetToken := &models.PasswordResetToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}

	// Save token to database
	err = s.resetTokenRepo.Create(ctx, resetToken)
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// TODO: Send password reset email with token
	// For now, we just store the token in the database
	// In production, this would send an email with a link containing the token

	return nil
}

// UpdatePassword validates reset token and updates user password
func (s *authService) UpdatePassword(ctx context.Context, token, newPassword string) error {
	// Get reset token from database
	resetToken, err := s.resetTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	// Check if token is already used
	if resetToken.Used {
		return errors.New("reset token has already been used")
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		return errors.New("reset token has expired")
	}

	// Hash new password with bcrypt cost 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update user password
	hashedPasswordStr := string(hashedPassword)
	user.PasswordHash = &hashedPasswordStr
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	err = s.resetTokenRepo.MarkAsUsed(ctx, resetToken.ID)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}
