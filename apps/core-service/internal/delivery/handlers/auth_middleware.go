package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"rekberkuy/core-service/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware acts as a security gate (RBAC) at the HTTP layer.
// The JWT secret is injected via the constructor so it does not read os.Getenv on
// every request and is easy to test.
type AuthMiddleware struct {
	jwtSecret string
}

// NewAuthMiddleware receives the JWT secret already validated by config.
func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

// RequireRole validates the real JWT token and checks role authorization.
func (a *AuthMiddleware) RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Access denied: Authorization header not found")
			c.Abort()
			return
		}

		// 2. Extract the token from the "Bearer <token>" format
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Access denied: Token format must be 'Bearer <token>'")
			c.Abort()
			return
		}

		// 3. Parse and validate the Token Claims using the struct from the user domain
		token, err := jwt.ParseWithClaims(tokenString, &domain.JWTCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			// Ensure the token encryption method is HMAC (HS256)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.jwtSecret), nil
		})

		// If the token is broken, tampered with, or expired, reject it immediately
		if err != nil || !token.Valid {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Access denied: Token invalid or expired")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*domain.JWTCustomClaims)
		if !ok {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Access denied: Failed to read claims payload")
			c.Abort()
			return
		}

		// 4. Match whether the Role in the Token is allowed to access this endpoint (RBAC)
		isAllowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			RespondError(c, http.StatusForbidden, "FORBIDDEN", "Permission denied: You are not authorized to execute this financial action!")
			c.Abort()
			return
		}

		// 5. Inject the verified User ID into the Gin Context
		// The purpose is so the usecase layer below knows who the transacting user is without re-parsing
		c.Set("user_id", claims.UserID)
		c.Set("user_role", string(claims.Role))

		c.Next()
	}
}

// RequirePersonalAccount ensures only personal accounts (USER) can access retail/consumer actions
// such as purchasing event tickets, strictly rejecting commercial accounts (EVENT_ORGANIZER, VENDOR, SELLER, SERVICE_PROVIDER).
func (a *AuthMiddleware) RequirePersonalAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("user_role")
		if !exists {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Access denied: User role not found in context")
			c.Abort()
			return
		}
		role := domain.UserRole(roleVal.(string))
		if domain.IsCommercialRole(role) {
			RespondError(c, http.StatusForbidden, "FORBIDDEN", "Permission denied: Commercial accounts (Event Organizer, Vendor, Seller, Service Provider) are strictly prohibited from purchasing event tickets. Only Personal Accounts (USER) are allowed.")
			c.Abort()
			return
		}
		c.Next()
	}
}
