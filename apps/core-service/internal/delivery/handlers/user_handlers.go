package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
)

// UserHandler exists for the dev-only token-test backdoor. Credential-based
// registration/login live on AuthHandler; this handler is intentionally small.
type UserHandler struct {
	userUsecase *usecase.UserUsecase
	jwtSecret   string
}

func NewUserHandler(uu *usecase.UserUsecase, jwtSecret string) *UserHandler {
	return &UserHandler{
		userUsecase: uu,
		jwtSecret:   jwtSecret,
	}
}

// CreateTokenTestRequest is the DEV-ONLY backdoor payload: it mints a JWT for an
// arbitrary user id / role without a password. The route is wired only outside production.
type CreateTokenTestRequest struct {
	UserID   string          `json:"user_id" binding:"required,uuid4"`
	Username string          `json:"username" binding:"required"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

// GenerateTokenTestHandler is a development convenience that issues a real JWT
// for any user id / role. Never registered in production.
func (h *UserHandler) GenerateTokenTestHandler(c *gin.Context) {
	var req CreateTokenTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtSecret := h.jwtSecret
	if jwtSecret == "" {
		jwtSecret = "rekberkuy-dev-secret-key"
	}

	claims := &domain.JWTCustomClaims{
		UserID:   req.UserID,
		Username: req.Username,
		Role:     req.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_type":   "Bearer",
		"access_token": tokenString,
	})
}
