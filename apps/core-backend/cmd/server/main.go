package main

import (
	"context"
	"fmt"
	"log"
	"os"


	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: Gagal memuat file .env, sistem akan membaca OS environment variables")
	}

	// 2. Ambil string koneksi database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Error: DATABASE_URL tidak ditemukan di file .env")
	}

	// 3. Inisialisasi Database Connection Pool (Skala Industri wajib pake Pool)
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Gagal memproses konfigurasi database: %v", err)
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Gagal membuat koneksi ke Supabase: %v", err)
	}
	defer dbPool.Close()

	// 4. Tes Koneksi (Ping)
	err = dbPool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Database Supabase tidak merespon: %v", err)
	}
	fmt.Println("🚀 HORE! Backend Go 1.25 berhasil terhubung dengan aman ke Database Supabase!")

	// 5. Inisialisasi Server Gin HTTP
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Rekberkuy Engine is running smoothly",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di port %s...", port)
	r.Run(":" + port)
}