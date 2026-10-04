package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"rekberkuy/core-service/internal/usecase"
)

// AuthHandler exposes credential-based registration and login.
type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthHandler(au *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: au}
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=4"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required"`
	Role        string `json:"role"`
	AccountType string `json:"account_type"`
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_INPUT", "invalid registration data: "+err.Error())
		return
	}

	user, err := h.authUsecase.Register(c.Request.Context(), req.Email, req.Username, req.Password, req.FullName, req.Role, req.AccountType)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailAlreadyRegistered) {
			RespondError(c, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered and RekberPay wallet created",
		"user":    user,
	})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid login data: " + err.Error()})
		return
	}

	token, user, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_type":   "Bearer",
		"access_token": token,
		"user":         user,
	})
}

func (h *AuthHandler) getMe(c *gin.Context) {
	
}