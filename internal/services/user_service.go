package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUserExists = errors.New("user already exists")
)

// UserService provides user management services
type UserService struct{}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{}
}

// Create creates a new user
func (s *UserService) Create(req *dto.CreateUserRequest, createdBy uuid.UUID) (*models.User, error) {
	// Check if user already exists
	var existing models.User
	if err := database.GetDB().Where("username = ? OR email = ?", req.Username, req.Email).First(&existing).Error; err == nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Default is_active to true if not specified
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	user := &models.User{
		Username:   req.Username,
		Password:   hashedPassword,
		ManagerID:  req.ManagerID,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		Email:      req.Email,
		Telephone:  req.Telephone,
		Role:       req.Role,
		IsActive:   isActive,
		ValidFrom:  req.ValidFrom,
		ValidTo:    req.ValidTo,
		VpnIP:      req.VpnIP,
		CreatedBy:  createdBy,
	}

	if err := database.GetDB().Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID gets a user by ID
func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.GetDB().Preload("Manager").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername gets a user by username
func (s *UserService) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := database.GetDB().Preload("Manager").First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (s *UserService) Update(id uuid.UUID, req *dto.UpdateUserRequest, updatedBy uuid.UUID) (*models.User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.MiddleName != "" {
		updates["middle_name"] = req.MiddleName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Telephone != "" {
		updates["telephone"] = req.Telephone
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.ManagerID != nil {
		updates["manager_id"] = req.ManagerID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.ValidFrom != nil {
		updates["valid_from"] = req.ValidFrom
	}
	if req.ValidTo != nil {
		updates["valid_to"] = req.ValidTo
	}
	if req.VpnIP != nil {
		updates["vpn_ip"] = *req.VpnIP
	}

	updates["updated_by"] = updatedBy

	if err := database.GetDB().Model(user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// UpdateProfile updates a user's own profile (limited fields)
func (s *UserService) UpdateProfile(id uuid.UUID, req *dto.UpdateProfileRequest, updatedBy uuid.UUID) (*models.User, error) {
	user, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.FirstName != "" {
		updates["first_name"] = req.FirstName
	}
	if req.MiddleName != "" {
		updates["middle_name"] = req.MiddleName
	}
	if req.LastName != "" {
		updates["last_name"] = req.LastName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Telephone != "" {
		updates["telephone"] = req.Telephone
	}

	updates["updated_by"] = updatedBy

	if err := database.GetDB().Model(user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(id uuid.UUID, currentPassword, newPassword string, updatedBy uuid.UUID) error {
	user, err := s.GetByID(id)
	if err != nil {
		return err
	}

	if !VerifyPassword(currentPassword, user.Password) {
		return ErrInvalidCredentials
	}

	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	return database.GetDB().Model(user).Updates(map[string]interface{}{
		"password":   hashedPassword,
		"updated_by": updatedBy,
	}).Error
}

// Delete soft deletes a user
func (s *UserService) Delete(id uuid.UUID) error {
	return database.GetDB().Delete(&models.User{}, "id = ?", id).Error
}

// List lists users with pagination
func (s *UserService) List(page, pageSize int, role models.Role, userID *uuid.UUID) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := database.GetDB().Model(&models.User{}).Preload("Manager")

	// Filter based on role
	switch role {
	case models.RoleAdmin:
		// Admins can see all users
	case models.RoleManager:
		// Managers can only see users they manage (their subordinates)
		if userID != nil {
			query = query.Where("manager_id = ?", userID)
		}
	case models.RoleUser:
		// Users can only see themselves
		if userID != nil {
			query = query.Where("id = ?", userID)
		}
	}

	// Count total
	query.Count(&total)

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetManagedUsers gets users managed by a specific manager
func (s *UserService) GetManagedUsers(managerID uuid.UUID, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := database.GetDB().Model(&models.User{}).Where("manager_id = ?", managerID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
