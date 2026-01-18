package repository

import (
	"errors"

	"github.com/joshleeeeee/LiteSSO/internal/database"
	"github.com/joshleeeeee/LiteSSO/internal/model"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	if err := database.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetByID finds a user by ID
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername finds a user by username
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail finds a user by email
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// ExistsByUsername checks if a user with the given username exists
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := database.DB.Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update updates a user
func (r *UserRepository) Update(user *model.User) error {
	return database.DB.Save(user).Error
}

// Delete soft-deletes a user
func (r *UserRepository) Delete(id uint) error {
	return database.DB.Delete(&model.User{}, id).Error
}
