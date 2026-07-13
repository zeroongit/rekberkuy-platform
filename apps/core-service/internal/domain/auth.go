package domain

import "github.com/golang-jwt/jwt/v5"

type JWTCustomClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
	jwt.RegisteredClaims
}