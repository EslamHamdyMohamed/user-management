package dtos

import "user-management/internal/models"

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email,max=255"`
	Password string `json:"password" binding:"omitempty,min=8,max=72"`
}
type UserListResponse struct {
	Users      []models.User `json:"users"`
	Pagination Pagination    `json:"pagination"`
}
