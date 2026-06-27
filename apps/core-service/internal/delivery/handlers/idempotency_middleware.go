package handlers

import (
	"bytes"
	"net/http"
	"rekberkuy/core-service/internal/domain"

	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// IdempotencyMiddleware memotong jalur request untuk menangkap token transaksi ganda
func IdempotencyMiddleware(repo domain.IdempotencyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Baca Idempotency-Key dari header
		idempotencyKey := c.GetHeader("Idempotency-Key")
		
		// Jika rute tidak mewajibkan atau key kosong, biarkan request lewat tanpa filter
		if idempotencyKey == "" {
			c.Next()
			return
		}

		// 2. Buat objek draf record awal
		record := &domain.IdempotencyRecord{
			ID:          idempotencyKey,
			RequestPath: c.Request.URL.Path,
		}

		// 3. Cek ke database secara atomik
		existingRecord, isNew, err := repo.CheckOrLock(c.Request.Context(), record)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sistem proteksi keuangan bermasalah: " + err.Error()})
			c.Abort()
			return
		}

		// 4. Jika TRANSKASI GANDA TERDETEKSI, bypass dan muntahkan response yang sama dari memori database!
		if !isNew {
			c.Header("X-Cache-Idempotency", "true")
			c.Data(existingRecord.ResponseStatus, "application/json", existingRecord.ResponseBody)
			c.Abort()
			return
		}

		// 5. Jika transaksi baru, jebak response body untuk dicatat hasilnya saat proses selesai
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		// 6. Sesaat setelah usecase selesai dieksekusi, perbarui kueri dengan response aslinya
		// (Fase MVP: mengasumsikan respons sukses langsung terekam aman)
	}
}