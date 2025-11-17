package models

import "time"

// User represents a user in the system
type User struct {
	ID            string     `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	PasswordHash  *string    `json:"-" db:"password_hash"`
	FirstName     *string    `json:"firstName" db:"first_name"`
	LastName      *string    `json:"lastName" db:"last_name"`
	OAuthProvider *string    `json:"oauthProvider" db:"oauth_provider"`
	OAuthID       *string    `json:"oauthId" db:"oauth_id"`
	CreatedAt     time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time  `json:"updatedAt" db:"updated_at"`
}
