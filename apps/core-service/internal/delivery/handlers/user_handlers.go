package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
	"os"
	"github.com/golang-jwt/jwt/v5"
	"time"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase // SUNTIKKAN INI
}

// Perbarui fungsi NewUserHandler agar menerima parameter usecase
func NewUserHandler(uu *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userUsecase: uu,
	}
}

type RegisterUserRequest struct {
	ID       string `json:"id" binding:"required,uuid4"`
	Username string `json:"username" binding:"required,min=4"`
	FullName string `json:"full_name" binding:"required"`
}

type CreateTokenTestRequest struct {
	UserID   string          `json:"user_id" binding:"required,uuid4"`
	Username string          `json:"username" binding:"required"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

func (h *UserHandler) GenerateTokenTestHandler(c *gin.Context) {
	var req CreateTokenTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rekberkuy-super-secret-key-fase-mvp"
	}

	claims := &domain.JWTCustomClaims{
		UserID:   req.UserID,
		Username: req.Username,
		Role:     req.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Aktif 1 hari
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) 
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate token: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_type":   "Bearer",
		"access_token": tokenString,
	})
}


func (h *UserHandler) RegisterProfileHandler(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data profil tidak valid: " + err.Error()})
		return
	}

	// Buat objek domain mapping dari request
	newUser := &domain.UserProfile{
		ID:       req.ID,
		Username: req.Username,
		FullName: req.FullName,
		Role:     domain.RoleUser, // Default MVP register
	}

	// Panggil usecase nyata untuk save profile & bikin wallet otomatis ke Supabase!
	err := h.userUsecase.RegisterNewUserProfile(c.Request.Context(), newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Profil pengguna dan dompet RekberPay berhasil diamankan!",
		"user_id": req.ID,
	})
}
