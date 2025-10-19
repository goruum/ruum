package guards

import (
	"strings"

	"github.com/goruum/ruum/core"
)

// AuthGuard checks for authentication
type AuthGuard struct {
	headerKey string
	tokenType string
}

// NewAuthGuard creates a new auth guard
func NewAuthGuard() *AuthGuard {
	return &AuthGuard{
		headerKey: "Authorization",
		tokenType: "Bearer",
	}
}

// WithHeaderKey sets the header key
func (g *AuthGuard) WithHeaderKey(key string) *AuthGuard {
	g.headerKey = key
	return g
}

// WithTokenType sets the token type
func (g *AuthGuard) WithTokenType(tokenType string) *AuthGuard {
	g.tokenType = tokenType
	return g
}

func (g *AuthGuard) CanActivate(ctx core.Context) (bool, error) {
	authHeader := ctx.Header(g.headerKey)
	
	if authHeader == "" {
		return false, core.UnauthorizedException("Missing authorization header")
	}
	
	// Extract token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != g.tokenType {
		return false, core.UnauthorizedException("Invalid authorization header format")
	}
	
	token := parts[1]
	
	// Store token in context for later use
	ctx.Set("auth_token", token)
	
	// In a real application, you would validate the token here
	// For now, we just check if it exists
	if token == "" {
		return false, core.UnauthorizedException("Invalid token")
	}
	
	return true, nil
}

// RolesGuard checks for required roles
type RolesGuard struct {
	requiredRoles []string
}

// NewRolesGuard creates a new roles guard
func NewRolesGuard(roles ...string) *RolesGuard {
	return &RolesGuard{
		requiredRoles: roles,
	}
}

func (g *RolesGuard) CanActivate(ctx core.Context) (bool, error) {
	// Get user roles from context (should be set by auth middleware)
	userRoles, ok := ctx.Get("user_roles").([]string)
	if !ok {
		return false, core.ForbiddenException("User roles not found")
	}
	
	// Check if user has any of the required roles
	for _, required := range g.requiredRoles {
		for _, userRole := range userRoles {
			if userRole == required {
				return true, nil
			}
		}
	}
	
	return false, core.ForbiddenException("Insufficient permissions")
}

