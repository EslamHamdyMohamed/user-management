package dtos

import (
	"user-management/internal/models"

	"github.com/google/uuid"
)

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email,max=255"`
	Password string `json:"password" binding:"omitempty,min=8,max=72"`
}
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Pagination Pagination     `json:"pagination"`
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func UserTransformer(user models.User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}
}

func UsersTransformer(users []models.User) []UserResponse {
	resp := make([]UserResponse, 0)
	for _, user := range users {
		resp = append(resp, UserTransformer(user))
	}
	return resp
}

func SafeUser(user *models.User) models.User {
	if user == nil {
		return models.User{}
	}
	return *user
}
