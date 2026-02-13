package auth

import (
	"net/http"
	"strings"

	"git.martianoids.com/martianoids/martian-stack/pkg/auth/jwt"
	"git.martianoids.com/martianoids/martian-stack/pkg/server/ctx"
)

// Middleware provides authentication middleware
type Middleware struct {
	jwtService *jwt.Service
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(jwtService *jwt.Service) *Middleware {
	return &Middleware{
		jwtService: jwtService,
	}
}

// RequireAuth is a middleware that requires authentication
func (m *Middleware) RequireAuth() ctx.Handler {
	return func(c ctx.Ctx) error {
		token := m.extractToken(c)
		if token == "" {
			return c.Error(http.StatusUnauthorized, "Missing authentication token")
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				return c.Error(http.StatusUnauthorized, "Token has expired")
			}
			return c.Error(http.StatusUnauthorized, "Invalid authentication token")
		}

		// Store user claims in context
		c.Store().Set("user", claims)
		c.Store().Set("user_id", claims.UserID)
		c.Store().Set("username", claims.Username)
		c.Store().Set("email", claims.Email)
		c.Store().Set("role", claims.Role)

		return c.Next()
	}
}

// RequireRole is a middleware that requires a specific role
func (m *Middleware) RequireRole(roles ...string) ctx.Handler {
	return func(c ctx.Ctx) error {
		// First, ensure user is authenticated
		var claims jwt.Claims
		if err := c.Store().Get("user", &claims); err != nil {
			return c.Error(http.StatusUnauthorized, "Not authenticated")
		}

		// Check if user has required role
		for _, role := range roles {
			if claims.Role == role {
				return c.Next()
			}
		}

		return c.Error(http.StatusForbidden, "Insufficient permissions")
	}
}

// OptionalAuth is a middleware that extracts auth but doesn't require it
func (m *Middleware) OptionalAuth() ctx.Handler {
	return func(c ctx.Ctx) error {
		token := m.extractToken(c)
		if token != "" {
			claims, err := m.jwtService.ValidateToken(token)
			if err == nil {
				// Store user claims in context
				c.Store().Set("user", claims)
				c.Store().Set("user_id", claims.UserID)
				c.Store().Set("username", claims.Username)
				c.Store().Set("email", claims.Email)
				c.Store().Set("role", claims.Role)
			}
		}

		return c.Next()
	}
}

// extractToken extracts the JWT token from the request
// Supports both Authorization header and cookie
func (m *Middleware) extractToken(c ctx.Ctx) string {
	// Try Authorization header first (Bearer token)
	authHeader := c.GetRequestHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try cookie as fallback
	token := c.GetCookie("access_token")
	if token != "" {
		return token
	}

	return ""
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(c ctx.Ctx) (*jwt.Claims, bool) {
	var claims jwt.Claims
	if err := c.Store().Get("user", &claims); err != nil {
		return nil, false
	}
	return &claims, true
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c ctx.Ctx) (string, bool) {
	userID := c.Store().GetString("user_id")
	return userID, userID != ""
}

// GetRoleFromContext extracts user role from context
func GetRoleFromContext(c ctx.Ctx) (string, bool) {
	role := c.Store().GetString("role")
	return role, role != ""
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c ctx.Ctx) bool {
	_, ok := GetUserFromContext(c)
	return ok
}

// HasRole checks if the current user has a specific role
func HasRole(c ctx.Ctx, role string) bool {
	userRole, ok := GetRoleFromContext(c)
	return ok && userRole == role
}

// HasAnyRole checks if the current user has any of the specified roles
func HasAnyRole(c ctx.Ctx, roles ...string) bool {
	userRole, ok := GetRoleFromContext(c)
	if !ok {
		return false
	}

	for _, role := range roles {
		if userRole == role {
			return true
		}
	}

	return false
}
