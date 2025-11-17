package handler

import (
	"errors"
	"net/http"

	"github.com/bipulkrdas/orgmind/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// SignUpRequest represents the request body for user signup
type SignUpRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

// SignUpResponse represents the response for successful signup
type SignUpResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse represents user data in API responses
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

// SignUp handles POST /api/auth/signup
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	user, token, err := h.authService.SignUp(c.Request.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, SignUpResponse{
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		},
		Token: token,
	})
}

// SignInRequest represents the request body for user signin
type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SignInResponse represents the response for successful signin
type SignInResponse struct {
	Token string `json:"token"`
}

// SignIn handles POST /api/auth/signin
func (h *AuthHandler) SignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	token, err := h.authService.SignIn(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sign in"})
		return
	}

	c.JSON(http.StatusOK, SignInResponse{
		Token: token,
	})
}

// ResetPasswordRequest represents the request body for password reset
type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPassword handles POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password reset"})
		return
	}

	// Always return success to prevent email enumeration
	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link has been sent"})
}

// UpdatePasswordRequest represents the request body for password update
type UpdatePasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// UpdatePassword handles POST /api/auth/update-password
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	err := h.authService.UpdatePassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// InitiateOAuth handles GET /api/auth/oauth/:provider
func (h *AuthHandler) InitiateOAuth(c *gin.Context) {
	provider := c.Param("provider")

	authURL, err := h.authService.InitiateOAuth(c.Request.Context(), provider)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported OAuth provider"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate OAuth"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"authUrl": authURL})
}

// OAuthCallbackRequest represents the query parameters for OAuth callback
type OAuthCallbackRequest struct {
	Code string `form:"code" binding:"required"`
}

// HandleOAuthCallback handles GET /api/auth/oauth/:provider/callback
func (h *AuthHandler) HandleOAuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	var req OAuthCallbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	token, err := h.authService.HandleOAuthCallback(c.Request.Context(), provider, req.Code)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedProvider) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported OAuth provider"})
			return
		}
		if errors.Is(err, service.ErrOAuthFailed) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth authentication failed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete OAuth"})
		return
	}

	c.JSON(http.StatusOK, SignInResponse{
		Token: token,
	})
}
