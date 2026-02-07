package service

import (
	"context"
	"errors"
	"fmt"
	"user-management/internal/dtos"
	"user-management/internal/models"
	"user-management/internal/repository"
	"user-management/internal/utils"

	"github.com/google/uuid"
)

type UserService interface {
	SignIn(ctx context.Context, email, password string) (*dtos.SignInResponse, error)
	GetUser(ctx context.Context, userID, targetID uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, userID, targetID uuid.UUID, req *dtos.UpdateUserRequest) (*models.User, error)
	ListUsers(ctx context.Context, lastID uuid.UUID, searchEmail string, limit int) ([]models.User, error)
	CreateUser(ctx context.Context, req *dtos.SignUpRequest) (*models.User, error)
}

type userService struct {
	repo            repository.UserRepository
	jwtManager      *utils.JWTManager
	passwordManager *utils.PasswordManager
}

func NewUserService(
	repo repository.UserRepository,
	jwtManager *utils.JWTManager,
	passwordManager *utils.PasswordManager,
) UserService {
	return &userService{
		repo:            repo,
		jwtManager:      jwtManager,
		passwordManager: passwordManager,
	}
}

func (s *userService) SignIn(ctx context.Context, email, password string) (*dtos.SignInResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := s.passwordManager.Compare(user.Password, password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dtos.SignInResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *userService) GetUser(ctx context.Context, userID, targetID uuid.UUID) (*models.User, error) {
	if userID != targetID {
		return nil, errors.New("unauthorized")
	}

	user, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *userService) UpdateUser(
	ctx context.Context,
	userID, targetID uuid.UUID,
	req *dtos.UpdateUserRequest,
) (*models.User, error) {
	// Strict privacy: only update self
	if userID != targetID {
		return nil, errors.New("unauthorized")
	}

	user, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update fields
	updateFields := make(map[string]interface{})

	if req.Email != "" && req.Email != user.Email {
		// Check email availability
		existingUser, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != targetID {
			return nil, errors.New("email already in use")
		}
		updateFields["email"] = req.Email
	}

	// Password update should ideally be separate, but if allowed here:
	if req.Password != "" {
		hashed, err := s.passwordManager.Hash(req.Password)
		if err != nil {
			return nil, err
		}
		updateFields["password"] = hashed
	}

	if len(updateFields) == 0 {
		return user, nil
	}

	// Perform update
	err = s.repo.Update(ctx, targetID, updateFields)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user
	updatedUser, err := s.repo.FindByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *userService) ListUsers(
	ctx context.Context,
	lastID uuid.UUID,
	searchEmail string,
	limit int,
) ([]models.User, error) {
	return s.repo.List(ctx, lastID, searchEmail, limit)
}

func (s *userService) CreateUser(ctx context.Context, req *dtos.SignUpRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := s.passwordManager.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
	}

	// Create user
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
