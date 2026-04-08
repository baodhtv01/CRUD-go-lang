package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/services"
	"github.com/baodhtv01/CRUD-go-lang/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, total, err := h.userService.GetAll(page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve users", err.Error())
		return
	}

	utils.PaginatedSuccessResponse(c, "users retrieved successfully", users, page, limit, total)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user ID", err.Error())
		return
	}

	// Non-admin users can only view their own profile
	callerID, role, err := getCallerContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "context error", err.Error())
		return
	}
	if role != models.RoleAdmin && callerID != id {
		utils.ErrorResponse(c, http.StatusForbidden, "access denied", "you can only view your own profile")
		return
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "user not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "user retrieved successfully", user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user ID", err.Error())
		return
	}

	callerID, role, err := getCallerContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "context error", err.Error())
		return
	}
	if role != models.RoleAdmin && callerID != id {
		utils.ErrorResponse(c, http.StatusForbidden, "access denied", "you can only update your own profile")
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Only admins can change roles
	if role != models.RoleAdmin {
		req.Role = ""
	}

	user, err := h.userService.Update(id, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "failed to update user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "user updated successfully", user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user ID", err.Error())
		return
	}

	callerID, role, err := getCallerContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "context error", err.Error())
		return
	}
	if role != models.RoleAdmin && callerID != id {
		utils.ErrorResponse(c, http.StatusForbidden, "access denied", "you can only delete your own account")
		return
	}

	if err := h.userService.Delete(id); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "failed to delete user", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "user deleted successfully", nil)
}

func parseUintParam(c *gin.Context, param string) (uint, error) {
	const maxUint = ^uint(0)
	val, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		return 0, err
	}
	if val > uint64(maxUint) {
		return 0, errors.New("ID value out of range")
	}
	return uint(val), nil
}

// getCallerContext safely retrieves userID and role from Gin context set by AuthMiddleware.
func getCallerContext(c *gin.Context) (uint, string, error) {
	rawID, exists := c.Get("userID")
	if !exists {
		return 0, "", errors.New("userID not found in context")
	}
	callerID, ok := rawID.(uint)
	if !ok {
		return 0, "", errors.New("userID has unexpected type in context")
	}
	rawRole, exists := c.Get("role")
	if !exists {
		return 0, "", errors.New("role not found in context")
	}
	role, ok := rawRole.(string)
	if !ok {
		return 0, "", errors.New("role has unexpected type in context")
	}
	return callerID, role, nil
}
