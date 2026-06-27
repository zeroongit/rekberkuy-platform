package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"
	"os"
	"strings"
	"fmt"
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

func AuthRoleMiddleware(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Header Authorization tidak ditemukan"})
			c.Abort()
			return
		}

		// 2. Ekstrak token dari format "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Format token harus 'Bearer <token>'"})
			c.Abort()
			return
		}

		// 3. Ambil JWT Secret dari environment/config
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "rekberkuy-super-secret-key-fase-mvp" // Fallback lokal
		}

		// 4. Parse dan validasi Token Claims
		token, err := jwt.ParseWithClaims(tokenString, &domain.JWTCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode signing tidak terduga: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak valid atau sudah kedaluwarsa: " + err.Error()})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*domain.JWTCustomClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Gagal membaca payload token"})
			c.Abort()
			return
		}

		// 5. Validasi Role-Based Access Control (RBAC)
		isAllowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Hak akses ditolak: Akun Anda tidak memiliki wewenang untuk rute finansial ini!"})
			c.Abort()
			return
		}

		// 6. Suntikkan User ID dan Data Claims ke Context Gin agar bisa dibaca oleh Usecase di baris bawah
		c.Set("user_id", claims.UserID)
		c.Set("user_role", string(claims.Role))

		c.Next()
	}
}