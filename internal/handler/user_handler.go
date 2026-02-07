package handler

import (
	"net/http"
	"strconv"
	"user-management/internal/dtos"
	"user-management/internal/service"
	"user-management/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   utils.ErrCodeInvalidID,
			Message: "Invalid user ID format",
		})
		return
	}

	// Get current user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Error:   utils.ErrCodeUnauthorized,
			Message: "User not authenticated",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Error:   utils.ErrCodeUnauthorized,
			Message: "Invalid session user ID",
		})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), userID, id)
	if err != nil {
		switch err.Error() {
		case "unauthorized":
			c.JSON(http.StatusForbidden, dtos.ErrorResponse{
				Error:   utils.ErrCodeForbidden,
				Message: "You don't have permission to view this user",
			})
		case "user not found":
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error:   utils.ErrCodeNotFound,
				Message: "User not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Error:   utils.ErrCodeInternalServerError,
				Message: "Failed to get user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dtos.UserTransformer(dtos.SafeUser(user)))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   utils.ErrCodeInvalidID,
			Message: "Invalid user ID format",
		})
		return
	}

	var req dtos.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   utils.ErrCodeValidationError,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	// Get current user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
			Error:   utils.ErrCodeUnauthorized,
			Message: "User not authenticated",
		})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error:   utils.ErrCodeInternalServerError,
			Message: "Invalid session",
		})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, id, &req)
	if err != nil {
		switch err.Error() {
		case "unauthorized":
			c.JSON(http.StatusForbidden, dtos.ErrorResponse{
				Error:   utils.ErrCodeForbidden,
				Message: err.Error(),
			})
		case "user not found":
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Error:   utils.ErrCodeNotFound,
				Message: "User not found",
			})
		case "email already in use":
			c.JSON(http.StatusConflict, dtos.ErrorResponse{
				Error:   utils.ErrCodeConflict,
				Message: "Email already in use",
			})
		default:
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Error:   utils.ErrCodeInternalServerError,
				Message: "Failed to update user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, dtos.UserTransformer(dtos.SafeUser(user)))
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	// Get query parameters
	lastIDStr := c.Query("last_id")
	searchEmail := c.Query("email")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	var lastID uuid.UUID
	if lastIDStr != "" {
		lastID, err = uuid.Parse(lastIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
				Error:   utils.ErrCodeInvalidCursor,
				Message: "Invalid last_id format",
			})
			return
		}
	}

	users, err := h.userService.ListUsers(c.Request.Context(), lastID, searchEmail, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Error:   utils.ErrCodeInternalServerError,
			Message: "Failed to list users",
		})
		return
	}

	var nextCursor string
	if len(users) > 0 {
		nextCursor = users[len(users)-1].ID.String()
	}

	c.JSON(http.StatusOK, dtos.UserListResponse{
		Users: dtos.UsersTransformer(users),
		Pagination: dtos.Pagination{
			Limit:      limit,
			NextCursor: nextCursor,
		},
	})
}
