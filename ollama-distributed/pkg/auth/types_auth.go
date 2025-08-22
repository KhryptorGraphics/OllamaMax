package auth

import "time"

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequestAuth represents a password change request (alternative)
type ChangePasswordRequestAuth struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// SessionResponse represents a session response (without sensitive data)
type SessionResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// UserResponse represents a user response (without sensitive data)
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRequestAuth represents a user update request (alternative)
type UpdateUserRequestAuth struct {
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}