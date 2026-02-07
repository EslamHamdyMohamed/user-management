package repository

import (
	"context"
	"errors"
	"user-management/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id uuid.UUID, updates interface{}) error
	List(ctx context.Context, lastID uuid.UUID, searchEmail string, limit int) ([]models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) Update(ctx context.Context, id uuid.UUID, updates interface{}) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *userRepository) List(ctx context.Context, lastID uuid.UUID, searchEmail string, limit int) ([]models.User, error) {
	var users []models.User
	query := r.db.WithContext(ctx).Model(&models.User{})

	if searchEmail != "" {
		query = query.Where("email ILIKE ?", "%"+searchEmail+"%")
	}

	if lastID != uuid.Nil {
		var lastUser models.User
		if err := r.db.WithContext(ctx).Select("created_at").First(&lastUser, "id = ?", lastID).Error; err != nil {
			return nil, err
		}
		// Fetch users created before the cursor, or same time but lower ID (assuming UUID logic or consistent sorts)
		// For standard pagination with sorting by CreatedAt DESC:
		// WHERE (created_at < last_created_at) OR (created_at = last_created_at AND id < last_id)
		query = query.Where("(created_at < ? OR (created_at = ? AND id < ?))", lastUser.CreatedAt, lastUser.CreatedAt, lastID)
	}

	err := query.Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}
