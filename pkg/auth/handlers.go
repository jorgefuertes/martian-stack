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
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AccountRepository defines the interface for account operations
type AccountRepository interface {
	GetByEmail(email string) (*adapter.Account, error)
	GetByUsername(username string) (*adapter.Account, error)
	Update(a *adapter.Account) error
}

// Handlers provides authentication HTTP handlers
type Handlers struct {
	repo       AccountRepository
	jwtService *jwt.Service
}

// NewHandlers creates new authentication handlers
func NewHandlers(repo AccountRepository, jwtService *jwt.Service) *Handlers {
	return &Handlers{
		repo:       repo,
		jwtService: jwtService,
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

		// Check if account is enabled
		if !account.Enabled {
			return c.Error(http.StatusForbidden, "Account is disabled")
		}

		// Validate password
		if err := account.ValidatePassword(req.Password); err != nil {
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

		refreshToken, err := h.jwtService.GenerateRefreshToken(account.ID)
		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate refresh token")
		}

		expiresAt, _ := h.jwtService.GetExpiryTime(accessToken)

		response := LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
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

// Refresh handles token refresh
func (h *Handlers) Refresh() ctx.Handler {
	return func(c ctx.Ctx) error {
		var req RefreshRequest
		if err := c.UnmarshalBody(&req); err != nil {
			return c.Error(http.StatusBadRequest, "Invalid request body")
		}

		if req.RefreshToken == "" {
			return c.Error(http.StatusBadRequest, "Refresh token is required")
		}

		// Validate refresh token
		claims, err := h.jwtService.ValidateToken(req.RefreshToken)
		if err != nil {
			return c.Error(http.StatusUnauthorized, "Invalid refresh token")
		}

		// Get account to generate new access token with current data
		var account *adapter.Account
		if claims.Email != "" {
			account, err = h.repo.GetByEmail(claims.Email)
			if err != nil {
				// Account not found, but we can still refresh with old claims
				account = nil
			}
		}

		// Generate new access token
		var accessToken string
		if account != nil {
			accessToken, err = h.jwtService.GenerateAccessToken(
				account.ID,
				account.Username,
				account.Email,
				account.Role,
			)
		} else {
			// Fallback: generate token with stored claims
			accessToken, err = h.jwtService.GenerateAccessToken(
				claims.UserID,
				claims.Username,
				claims.Email,
				claims.Role,
			)
		}

		if err != nil {
			return c.Error(http.StatusInternalServerError, "Failed to generate token")
		}

		expiresAt, _ := h.jwtService.GetExpiryTime(accessToken)

		response := RefreshResponse{
			AccessToken: accessToken,
			ExpiresAt:   expiresAt,
		}

		return c.SendJSON(response)
	}
}

// Logout handles user logout
func (h *Handlers) Logout() ctx.Handler {
	return func(c ctx.Ctx) error {
		// In a stateless JWT system, logout is handled client-side by removing the token
		// For enhanced security, you could:
		// 1. Add the token to a blacklist (requires storage)
		// 2. Use short-lived tokens with refresh tokens
		// 3. Implement token revocation

		// For now, return success
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

		// Get fresh user data
		account, err := h.repo.GetByEmail(claims.Email)
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
