package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/domain"
	"rekberkuy/core-service/internal/usecase"

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

		userRole := c.GetHeader("X-User-Role")
		
		if userRole == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token autentikasi atau rahasia role tidak ditemukan"})
			c.Abort()
			return
		}

		isAllowed := false
		for _, role := range allowedRoles {
			if string(role) == userRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Hak akses ditolak: Anda tidak memiliki wewenang untuk mengeksekusi aksi finansial ini!"})
			c.Abort()
			return
		}

		c.Next()
	}
}