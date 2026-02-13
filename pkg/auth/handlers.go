package auth

import (
	"net/http"
	"time"

	"git.martianoids.com/martianoids/martian-stack/pkg/auth/jwt"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/adapter"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
)

// LoginRequest represents a login request payload
type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents user information returned after login
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshResponse represents a token refresh response
type RefreshResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AccountRepository defines the interface for account operations
type AccountRepository interface {
	GetByEmail(email string) (*adapter.Account, error)
	GetByUsername(username string) (*adapter.Account, error)
	Get(id string) (*adapter.Account, error)
	Update(a *adapter.Account) error
}

// Handlers provides authentication HTTP handlers
type Handlers struct {
	repo              AccountRepository
	jwtService        *jwt.Service
	refreshTokenRepo  adapter.RefreshTokenRepository
	resetTokenRepo    adapter.PasswordResetTokenRepository
}

// NewHandlers creates new authentication handlers
func NewHandlers(
	repo AccountRepository,
	jwtService *jwt.Service,
	refreshTokenRepo adapter.RefreshTokenRepository,
	resetTokenRepo adapter.PasswordResetTokenRepository,
) *Handlers {
	return &Handlers{
		repo:             repo,
		jwtService:       jwtService,
		refreshTokenRepo: refreshTokenRepo,
		resetTokenRepo:   resetTokenRepo,
	}
}

// Login handles user login
func (h *Handlers) Login() ctx.Handler {
	return func(c ctx.Ctx) error {
		var req LoginRequest
		if err := c.UnmarshalBody(&req); err != nil {
			return c.Error(http.StatusBadRequest, "Invalid request body")
		}

		// Validate request
		if req.Email == "" && req.Username == "" {
			return c.Error(http.StatusBadRequest, "Email or username is required")
		}

		if req.Password == "" {
			return c.Error(http.StatusBadRequest, "Password is required")
		}

		// Get account
		var account *adapter.Account
		var err error

		if req.Email != "" {
			account, err = h.repo.GetByEmail(req.Email)
		} else {
			account, err = h.repo.GetByUsername(req.Username)
		}

		if err != nil {
			return c.Error(http.StatusUnauthorized, "Invalid credentials")
		}

		// Validate password and check if account is enabled
		// Use the same error message for all failures to prevent account enumeration
		if err := account.ValidatePassword(req.Password); err != nil {
			return c.Error(http.StatusUnauthorized, "Invalid credentials")
		}

		if !account.Enabled {
			return c.Error(http.StatusUnauthorized, "Invalid credentials")
		}

		// Update last login
		account.LastLogin = time.Now()
		if err := h.repo.Update(account); err != nil {
			// Log error but don't fail login
		}

		// Generate tokens
		accessToken, err := h.jwtService.GenerateAccessToken(
			account.ID,
			account.Username,
			account.Email,
			account.Role,
		)
		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate token")
		}

		// Generate refresh token (raw token to send to user)
		rawRefreshToken, tokenHash, err := adapter.GenerateSecureToken()
		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate refresh token")
		}

		// Store refresh token in database
		refreshTokenExpiry := 7 * 24 * time.Hour // 7 days
		refreshTokenRecord := adapter.NewRefreshToken(account.ID, tokenHash, refreshTokenExpiry)
		if err := h.refreshTokenRepo.Create(refreshTokenRecord); err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to store refresh token")
		}

		expiresAt, _ := h.jwtService.GetExpiryTime(accessToken)

		response := LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: rawRefreshToken,
			ExpiresAt:    expiresAt,
			User: UserInfo{
				ID:       account.ID,
				Username: account.Username,
				Email:    account.Email,
				Name:     account.Name,
				Role:     account.Role,
			},
		}

		return c.SendJSON(response)
	}
}

// Refresh handles token refresh with rotation
func (h *Handlers) Refresh() ctx.Handler {
	return func(c ctx.Ctx) error {
		var req RefreshRequest
		if err := c.UnmarshalBody(&req); err != nil {
			return c.Error(http.StatusBadRequest, "Invalid request body")
		}

		if req.RefreshToken == "" {
			return c.Error(http.StatusBadRequest, "Refresh token is required")
		}

		// Hash the provided token to look it up in the database
		tokenHash, err := adapter.HashToken(req.RefreshToken)
		if err != nil {
			return c.Error(http.StatusBadRequest, "Invalid refresh token format")
		}

		// Get refresh token from database
		storedToken, err := h.refreshTokenRepo.GetByTokenHash(tokenHash)
		if err != nil {
			return c.Error(http.StatusUnauthorized, "Invalid refresh token")
		}

		// Validate token
		if !storedToken.IsValid() {
			if storedToken.IsExpired() {
				return c.Error(http.StatusUnauthorized, "Refresh token expired")
			}
			if storedToken.IsRevoked() {
				return c.Error(http.StatusUnauthorized, "Refresh token revoked")
			}
			return c.Error(http.StatusUnauthorized, "Invalid refresh token")
		}

		// Get account to generate new tokens with current data
		account, err := h.repo.Get(storedToken.UserID)
		if err != nil {
			return c.Error(http.StatusUnauthorized, "Account not found")
		}

		// Check if account is enabled
		if !account.Enabled {
			return c.Error(http.StatusForbidden, "Account is disabled")
		}

		// Revoke the old refresh token (token rotation)
		if err := h.refreshTokenRepo.Revoke(tokenHash); err != nil {
			// Log error but continue - we don't want to fail the refresh
		}

		// Generate new access token
		accessToken, err := h.jwtService.GenerateAccessToken(
			account.ID,
			account.Username,
			account.Email,
			account.Role,
		)
		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate token")
		}

		// Generate new refresh token
		rawRefreshToken, newTokenHash, err := adapter.GenerateSecureToken()
		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate refresh token")
		}

		// Store new refresh token in database
		refreshTokenExpiry := 7 * 24 * time.Hour // 7 days
		newRefreshTokenRecord := adapter.NewRefreshToken(account.ID, newTokenHash, refreshTokenExpiry)
		if err := h.refreshTokenRepo.Create(newRefreshTokenRecord); err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to store refresh token")
		}

		expiresAt, _ := h.jwtService.GetExpiryTime(accessToken)

		response := RefreshResponse{
			AccessToken:  accessToken,
			RefreshToken: rawRefreshToken,
			ExpiresAt:    expiresAt,
		}

		return c.SendJSON(response)
	}
}

// Logout handles user logout
func (h *Handlers) Logout() ctx.Handler {
	return func(c ctx.Ctx) error {
		// Get user from context (set by auth middleware)
		var claims jwt.Claims
		if err := c.Store().Get("user", &claims); err != nil {
			// Not authenticated, but that's okay - just return success
			return c.SendJSON(map[string]string{
				"message": "Logged out successfully",
			})
		}

		// Revoke all refresh tokens for this user
		if err := h.refreshTokenRepo.RevokeAll(claims.UserID); err != nil {
			// Log error but don't fail logout
		}

		// Client should remove the token from storage
		return c.SendJSON(map[string]string{
			"message": "Logged out successfully",
		})
	}
}

// Me returns the current authenticated user's information
func (h *Handlers) Me() ctx.Handler {
	return func(c ctx.Ctx) error {
		// Get user from context (set by auth middleware)
		var claims jwt.Claims
		if err := c.Store().Get("user", &claims); err != nil {
			return c.Error(http.StatusUnauthorized, "Not authenticated")
		}

		// Get fresh user data by ID (immutable identifier)
		account, err := h.repo.Get(claims.UserID)
		if err != nil {
			return c.Error(http.StatusNotFound, "User not found")
		}

		userInfo := UserInfo{
			ID:       account.ID,
			Username: account.Username,
			Email:    account.Email,
			Name:     account.Name,
			Role:     account.Role,
		}

		return c.SendJSON(userInfo)
	}
}

// PasswordResetRequestRequest represents a password reset request payload
type PasswordResetRequestRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetRequestResponse represents the response to a password reset request
type PasswordResetRequestResponse struct {
	Message string `json:"message"`
}

// RequestPasswordReset handles password reset requests
func (h *Handlers) RequestPasswordReset() ctx.Handler {
	return func(c ctx.Ctx) error {
		var req PasswordResetRequestRequest
		if err := c.UnmarshalBody(&req); err != nil {
			return c.Error(http.StatusBadRequest, "Invalid request body")
		}

		if req.Email == "" {
			return c.Error(http.StatusBadRequest, "Email is required")
		}

		// Get account by email
		account, err := h.repo.GetByEmail(req.Email)
		if err != nil {
			// For security, always return success even if email doesn't exist
			// This prevents email enumeration attacks
			return c.SendJSON(PasswordResetRequestResponse{
				Message: "If an account with that email exists, a password reset link has been sent",
			})
		}

		// Check if account is enabled
		if !account.Enabled {
			// For security, return success even if account is disabled
			return c.SendJSON(PasswordResetRequestResponse{
				Message: "If an account with that email exists, a password reset link has been sent",
			})
		}

		// Delete any existing password reset tokens for this user
		if err := h.resetTokenRepo.DeleteByUserID(account.ID); err != nil {
			// Log error but continue
		}

		// Generate password reset token
		rawToken, tokenHash, err := adapter.GenerateSecureToken()
		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate reset token")
		}

		// Store password reset token (valid for 1 hour)
		resetTokenExpiry := 1 * time.Hour
		resetToken := adapter.NewPasswordResetToken(account.ID, tokenHash, resetTokenExpiry)
		if err := h.resetTokenRepo.Create(resetToken); err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to store reset token")
		}

		// TODO: Send email with reset link containing rawToken
		// The rawToken should be included in the email link, never in the API response
		_ = rawToken

		return c.SendJSON(PasswordResetRequestResponse{
			Message: "If an account with that email exists, a password reset link has been sent",
		})
	}
}

// PasswordResetRequest represents a password reset payload
type PasswordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// PasswordResetResponse represents the response to a password reset
type PasswordResetResponse struct {
	Message string `json:"message"`
}

// ResetPassword handles password reset using a valid token
func (h *Handlers) ResetPassword() ctx.Handler {
	return func(c ctx.Ctx) error {
		var req PasswordResetRequest
		if err := c.UnmarshalBody(&req); err != nil {
			return c.Error(http.StatusBadRequest, "Invalid request body")
		}

		if req.Token == "" {
			return c.Error(http.StatusBadRequest, "Token is required")
		}

		if req.NewPassword == "" {
			return c.Error(http.StatusBadRequest, "New password is required")
		}

		// Hash the provided token to look it up in the database
		tokenHash, err := adapter.HashToken(req.Token)
		if err != nil {
			return c.Error(http.StatusBadRequest, "Invalid token format")
		}

		// Get password reset token from database
		storedToken, err := h.resetTokenRepo.GetByTokenHash(tokenHash)
		if err != nil {
			return c.Error(http.StatusUnauthorized, "Invalid or expired reset token")
		}

		// Validate token
		if !storedToken.IsValid() {
			if storedToken.IsExpired() {
				return c.Error(http.StatusUnauthorized, "Reset token has expired")
			}
			if storedToken.IsUsed() {
				return c.Error(http.StatusUnauthorized, "Reset token has already been used")
			}
			return c.Error(http.StatusUnauthorized, "Invalid reset token")
		}

		// Get account
		account, err := h.repo.Get(storedToken.UserID)
		if err != nil {
			return c.Error(http.StatusNotFound, "Account not found")
		}

		// Set new password
		if err := account.SetPassword(req.NewPassword); err != nil {
			return c.Error(http.StatusBadRequest, err.Error())
		}

		// Update account
		if err := h.repo.Update(account); err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to update password")
		}

		// Mark token as used
		if err := h.resetTokenRepo.MarkAsUsed(tokenHash); err != nil {
			// Log error but don't fail the reset
		}

		// Revoke all refresh tokens to force re-login
		if err := h.refreshTokenRepo.RevokeAll(account.ID); err != nil {
			// Log error but don't fail the reset
		}

		return c.SendJSON(PasswordResetResponse{
			Message: "Password has been reset successfully",
		})
	}
}
