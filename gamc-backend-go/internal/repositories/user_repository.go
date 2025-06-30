// internal/repositories/user_repository.go
package repositories

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository maneja las operaciones de base de datos para usuarios
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository crea una nueva instancia del repositorio de usuarios
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create crea un nuevo usuario
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID obtiene un usuario por ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Preload("SecurityQuestions", "is_active = ?", true).
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail obtiene un usuario por email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Preload("SecurityQuestions", "is_active = ?", true).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername obtiene un usuario por username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update actualiza un usuario
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdatePartial actualiza campos específicos de un usuario
func (r *UserRepository) UpdatePartial(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete elimina un usuario (soft delete - desactivar)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// HardDelete elimina permanentemente un usuario
func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.User{}).Error
}

// GetByFilter obtiene usuarios con filtros y paginación
func (r *UserRepository) GetByFilter(ctx context.Context, filter *UserFilter) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})

	// Aplicar filtros
	query = r.applyUserFilters(query, filter)

	// Contar total antes de paginar
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar ordenamiento
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortDesc {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// Aplicar paginación
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Precargar relaciones
	query = query.Preload("OrganizationalUnit")

	// Ejecutar consulta
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetByRole obtiene usuarios por rol
func (r *UserRepository) GetByRole(ctx context.Context, role string) ([]*models.User, error) {
	var users []*models.User
	err := r.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Where("role = ? AND is_active = ?", role, true).
		Find(&users).Error
	return users, err
}

// GetByOrganizationalUnit obtiene usuarios por unidad organizacional
func (r *UserRepository) GetByOrganizationalUnit(ctx context.Context, unitID int) ([]*models.User, error) {
	var users []*models.User
	err := r.db.WithContext(ctx).
		Where("organizational_unit_id = ? AND is_active = ?", unitID, true).
		Find(&users).Error
	return users, err
}

// UpdateLastLogin actualiza la última fecha de login
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("last_login", now).Error
}

// UpdatePassword actualiza la contraseña de un usuario
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"password_hash":       passwordHash,
			"password_changed_at": time.Now(),
		}).Error
}

// ExistsByEmail verifica si existe un usuario con el email dado
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

// ExistsByUsername verifica si existe un usuario con el username dado
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}

// GetActiveUserCount obtiene el conteo de usuarios activos
func (r *UserRepository) GetActiveUserCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("is_active = ?", true).
		Count(&count).Error
	return count, err
}

// GetUserStats obtiene estadísticas de usuarios
func (r *UserRepository) GetUserStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{}

	// Total usuarios
	r.db.WithContext(ctx).Model(&models.User{}).Count(&stats.Total)

	// Usuarios activos
	r.db.WithContext(ctx).Model(&models.User{}).Where("is_active = ?", true).Count(&stats.Active)

	// Por rol
	var roleCounts []struct {
		Role  string
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("role, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("role").
		Scan(&roleCounts)

	stats.ByRole = make(map[string]int64)
	for _, rc := range roleCounts {
		stats.ByRole[rc.Role] = rc.Count
	}

	// Por unidad organizacional
	var unitCounts []struct {
		UnitID int
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("organizational_unit_id as unit_id, COUNT(*) as count").
		Where("is_active = ? AND organizational_unit_id IS NOT NULL", true).
		Group("organizational_unit_id").
		Scan(&unitCounts)

	stats.ByUnit = make(map[int]int64)
	for _, uc := range unitCounts {
		stats.ByUnit[uc.UnitID] = uc.Count
	}

	// Últimos registros (últimos 30 días)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&stats.RecentRegistrations)

	return stats, nil
}

// SearchUsers busca usuarios por nombre o email
func (r *UserRepository) SearchUsers(ctx context.Context, searchTerm string, limit int) ([]*models.User, error) {
	var users []*models.User
	searchPattern := fmt.Sprintf("%%%s%%", searchTerm)

	query := r.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Where("is_active = ?", true).
		Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR username ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&users).Error
	return users, err
}

// GetInactiveUsers obtiene usuarios inactivos por más de X días
func (r *UserRepository) GetInactiveUsers(ctx context.Context, days int) ([]*models.User, error) {
	var users []*models.User
	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := r.db.WithContext(ctx).
		Where("last_login < ? OR last_login IS NULL", cutoffDate).
		Where("is_active = ?", true).
		Find(&users).Error

	return users, err
}

// applyUserFilters aplica los filtros a la consulta de usuarios
func (r *UserRepository) applyUserFilters(query *gorm.DB, filter *UserFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Filtro por estado activo
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	// Filtro por rol
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	// Filtro por unidad organizacional
	if filter.OrganizationalUnitID != nil {
		query = query.Where("organizational_unit_id = ?", *filter.OrganizationalUnitID)
	}

	// Búsqueda por texto
	if filter.SearchTerm != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.SearchTerm)
		query = query.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR username ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Filtro por fecha de creación
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filter.CreatedTo)
	}

	// Filtro por último login
	if filter.LastLoginFrom != nil {
		query = query.Where("last_login >= ?", *filter.LastLoginFrom)
	}
	if filter.LastLoginTo != nil {
		query = query.Where("last_login <= ?", *filter.LastLoginTo)
	}

	return query
}

// UserFilter estructura para filtrar usuarios
type UserFilter struct {
	IsActive             *bool
	Role                 string
	OrganizationalUnitID *int
	SearchTerm           string
	CreatedFrom          *time.Time
	CreatedTo            *time.Time
	LastLoginFrom        *time.Time
	LastLoginTo          *time.Time
	SortBy               string
	SortDesc             bool
	Limit                int
	Offset               int
}

// UserStats estadísticas de usuarios
type UserStats struct {
	Total               int64
	Active              int64
	ByRole              map[string]int64
	ByUnit              map[int]int64
	RecentRegistrations int64
}
