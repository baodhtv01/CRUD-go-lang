package handlers

import (
	"net/http"

	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/services"
	"github.com/baodhtv01/CRUD-go-lang/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusConflict, "registration failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "user registered successfully", user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "login failed", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "login successful", resp)
}
