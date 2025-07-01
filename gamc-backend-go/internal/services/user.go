package services

import (
	"context"
	"errors"
	"fmt"

	"gamc-backend-go/internal/database/models"

	//"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService maneja la lógica de negocio de usuarios
type UserService struct {
	db *gorm.DB
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// GetUserByID obtiene un usuario por ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Where("id = ? AND is_active = ?", userID, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	return &user, nil
}

// GetUserByEmail obtiene un usuario por email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	return &user, nil
}
