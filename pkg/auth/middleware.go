package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides JWT authentication middleware for Gin
type AuthMiddleware struct {
	jwtService *JWTService
	rbac       *RBAC
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *JWTService, rbac *RBAC) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		rbac:       rbac,
	}
}

// RequireAuth middleware that requires valid JWT authentication
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
				"code":  "AUTH_TOKEN_MISSING",
			})
			c.Abort()
			return
		}

		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "AUTH_TOKEN_INVALID",
			})
			c.Abort()
			return
		}

		// Check if user is active
		user, err := am.rbac.GetUser(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
				"code":  "AUTH_USER_NOT_FOUND",
			})
			c.Abort()
			return
		}

		if !user.Active {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User account is inactive",
				"code":  "AUTH_USER_INACTIVE",
			})
			c.Abort()
			return
		}

		// Store claims in context for use in handlers
		c.Set("claims", claims)
		c.Set("user", user)
		c.Next()
	}
}

// RequirePermission middleware that requires specific permission
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure authentication
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication context not found",
				"code":  "AUTH_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*Claims)
		hasPermission, err := am.rbac.HasPermission(userClaims.UserID, permission)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Permission check failed",
				"code":  "AUTH_PERMISSION_CHECK_FAILED",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "Insufficient permissions",
				"code":       "AUTH_INSUFFICIENT_PERMISSIONS",
				"required":   permission,
				"user_role":  userClaims.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (am *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure authentication
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication context not found",
				"code":  "AUTH_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*Claims)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "User context not found",
				"code":  "AUTH_USER_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		userData := user.(*User)
		hasRole := false
		for _, userRole := range userData.Roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "Insufficient role",
				"code":        "AUTH_INSUFFICIENT_ROLE",
				"required":    role,
				"user_roles":  userData.Roles,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (am *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return am.RequireRole(RoleAdmin)
}

// RequireOperator middleware that requires operator role or higher
func (am *AuthMiddleware) RequireOperator() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First ensure authentication
		am.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Authentication context not found",
				"code":  "AUTH_CONTEXT_MISSING",
			})
			c.Abort()
			return
		}

		userClaims := claims.(*Claims)
		if !userClaims.IsOperator() {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Operator role or higher required",
				"code":      "AUTH_INSUFFICIENT_ROLE",
				"user_role": userClaims.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that extracts auth info if present but doesn't require it
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := am.jwtService.ValidateToken(token)
		if err != nil {
			// Invalid token, but we don't abort for optional auth
			c.Next()
			return
		}

		// Check if user exists and is active
		user, err := am.rbac.GetUser(claims.UserID)
		if err != nil || !user.Active {
			c.Next()
			return
		}

		// Store claims in context
		c.Set("claims", claims)
		c.Set("user", user)
		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// GetCurrentUser helper function to get current user from context
func GetCurrentUser(c *gin.Context) (*User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	userData, ok := user.(*User)
	return userData, ok
}

// GetCurrentClaims helper function to get current claims from context
func GetCurrentClaims(c *gin.Context) (*Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	claimsData, ok := claims.(*Claims)
	return claimsData, ok
}

// HasPermissionInContext checks if current user has permission
func HasPermissionInContext(c *gin.Context, permission string, rbac *RBAC) bool {
	claims, exists := GetCurrentClaims(c)
	if !exists {
		return false
	}

	hasPermission, err := rbac.HasPermission(claims.UserID, permission)
	if err != nil {
		return false
	}

	return hasPermission
}

// IsAdminInContext checks if current user is admin
func IsAdminInContext(c *gin.Context) bool {
	claims, exists := GetCurrentClaims(c)
	if !exists {
		return false
	}
	return claims.IsAdmin()
}

// IsOperatorInContext checks if current user is operator or higher
func IsOperatorInContext(c *gin.Context) bool {
	claims, exists := GetCurrentClaims(c)
	if !exists {
		return false
	}
	return claims.IsOperator()
}
