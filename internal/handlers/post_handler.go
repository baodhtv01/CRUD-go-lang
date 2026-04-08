package handlers

import (
	"net/http"
	"strconv"

	"github.com/baodhtv01/CRUD-go-lang/internal/models"
	"github.com/baodhtv01/CRUD-go-lang/internal/services"
	"github.com/baodhtv01/CRUD-go-lang/internal/utils"
	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postService services.PostService
}

func NewPostHandler(postService services.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, total, err := h.postService.GetAll(page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve posts", err.Error())
		return
	}

	utils.PaginatedSuccessResponse(c, "posts retrieved successfully", posts, page, limit, total)
}

func (h *PostHandler) GetByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid post ID", err.Error())
		return
	}

	post, err := h.postService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "post not found", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "post retrieved successfully", post)
}

func (h *PostHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req models.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	post, err := h.postService.Create(userID.(uint), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to create post", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "post created successfully", post)
}

func (h *PostHandler) Update(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid post ID", err.Error())
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var req models.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	post, err := h.postService.Update(id, userID.(uint), role.(string), &req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "post not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "forbidden: you can only update your own posts" {
			statusCode = http.StatusForbidden
		}
		utils.ErrorResponse(c, statusCode, "failed to update post", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "post updated successfully", post)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid post ID", err.Error())
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	if err := h.postService.Delete(id, userID.(uint), role.(string)); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "post not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "forbidden: you can only delete your own posts" {
			statusCode = http.StatusForbidden
		}
		utils.ErrorResponse(c, statusCode, "failed to delete post", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "post deleted successfully", nil)
}
