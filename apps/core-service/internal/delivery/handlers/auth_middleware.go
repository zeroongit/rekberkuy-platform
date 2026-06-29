package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"rekberkuy/core-service/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthRoleMiddleware bertindak sebagai gerbang pengaman (RBAC) di layer HTTP.
// Middleware ini memotong request, memvalidasi token JWT asli, dan memeriksa wewenang role.
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

		// 3. Ambil JWT Secret dari environment
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "rekberkuy-super-secret-key-fase-mvp" // Fallback aman untuk development lokal
		}

		// 4. Parse dan validasi Token Claims menggunakan struct dari domain user
		token, err := jwt.ParseWithClaims(tokenString, &domain.JWTCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
			// Pastikan metode enkripsi token adalah HMAC (HS256)
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("metode signing tidak terduga: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		// Jika token rusak, dimanipulasi, atau sudah expired, langsung tendang balik
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Token tidak valid atau sudah kedaluwarsa"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*domain.JWTCustomClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak: Gagal membaca payload claims"})
			c.Abort()
			return
		}

		// 5. Cocokkan apakah Role yang ada di Token diizinkan mengakses endpoint ini (RBAC)
		isAllowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Hak akses ditolak: Anda tidak memiliki wewenang untuk mengeksekusi aksi finansial ini!"})
			c.Abort()
			return
		}

		// 6. Suntikkan User ID hasil verifikasi aman ke dalam Context Gin
		// Tujuannya agar layer usecase di bawahnya bisa tahu siapa user yang sedang bertransaksi tanpa perlu parsing ulang
		c.Set("user_id", claims.UserID)
		c.Set("user_role", string(claims.Role))

		c.Next()
	}
}