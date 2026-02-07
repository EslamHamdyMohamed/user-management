package handler

import (
	"net/http"
	"strings"
	"user-management/internal/dtos"
	"user-management/internal/service"
	"user-management/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService service.UserService
	jwtManager  *utils.JWTManager
}

func NewAuthHandler(userService service.UserService, jwtManager *utils.JWTManager) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtManager:  jwtManager,
	}
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req dtos.SignInRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   utils.ErrCodeValidationError,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	// Trim email
	req.Email = strings.TrimSpace(req.Email)
	req.Email = strings.ToLower(req.Email)

	response, err := h.userService.SignIn(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		message := "Authentication failed"

		if strings.Contains(err.Error(), "disabled") {
			message = "Account is disabled"
		}

		c.JSON(status, dtos.ErrorResponse{
			Error:   utils.ErrCodeAuthFailed,
			Message: message,
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req dtos.SignUpRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Error:   utils.ErrCodeValidationError,
			Message: "Invalid request payload",
			Details: err.Error(),
		})
		return
	}

	// Trim email
	req.Email = strings.TrimSpace(req.Email)
	req.Email = strings.ToLower(req.Email)

	_, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusBadRequest
		}

		c.JSON(status, dtos.ErrorResponse{
			Error:   utils.ErrCodeSignupFailed,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.SuccessResponse{
		Message: "User created successfully",
	})
}
