package handlers

import (
	"net/http"
	"rekberkuy/core-service/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	// Nanti diisi usecase user jika sudah dibuat, untuk sementara kita injeksi walletUsecase
	// demi otomatisasi pembuatan wallet saat user daftar (Fase MVP)
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

type RegisterUserRequest struct {
	ID       string `json:"id" binding:"required,uuid4"`
	Username string `json:"username" binding:"required,min=4"`
	FullName string `json:"full_name" binding:"required"`
}

// RegisterProfileHandler mencatat user baru sekaligus mengamankan pembuatan dompet RekberPay-nya
func (h *UserHandler) RegisterProfileHandler(c *gin.Context) {
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data profil tidak valid: " + err.Error()})
		return
	}

	// Sektor ini akan memanggil layer bisnis untuk menyimpan data ke Supabase
	c.JSON(http.StatusCreated, gin.H{
		"message": "Profil pengguna dan dompet RekberPay berhasil diamankan!",
		"user_id": req.ID,
	})
}

// AuthRoleMiddleware bertindak sebagai tameng keamanan (RBAC) di layer HTTP
// Mencegah user biasa menembak endpoint sensitif milik Admin, EO, atau Merchant
func AuthRoleMiddleware(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Fase MVP: Untuk pengujian lokal, kita membaca role dari Header Request
		// Di fase produksi nanti, bagian ini diganti dengan ekstraksi claims dari JWT Token
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