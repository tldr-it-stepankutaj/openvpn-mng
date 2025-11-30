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
	ErrGroupNotFound = errors.New("group not found")
	ErrGroupExists   = errors.New("group already exists")
)

// GroupService provides group management services
type GroupService struct{}

// NewGroupService creates a new group service
func NewGroupService() *GroupService {
	return &GroupService{}
}

// Create creates a new group
func (s *GroupService) Create(req *dto.CreateGroupRequest, createdBy uuid.UUID) (*models.Group, error) {
	// Check if group already exists
	var existing models.Group
	if err := database.GetDB().Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, ErrGroupExists
	}

	group := &models.Group{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   createdBy,
	}

	if err := database.GetDB().Create(group).Error; err != nil {
		return nil, err
	}

	return group, nil
}

// GetByID gets a group by ID
func (s *GroupService) GetByID(id uuid.UUID) (*models.Group, error) {
	var group models.Group
	if err := database.GetDB().First(&group, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}
	return &group, nil
}

// Update updates a group
func (s *GroupService) Update(id uuid.UUID, req *dto.UpdateGroupRequest, updatedBy uuid.UUID) (*models.Group, error) {
	group, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	updates["updated_by"] = updatedBy

	if err := database.GetDB().Model(group).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Delete soft deletes a group
func (s *GroupService) Delete(id uuid.UUID) error {
	return database.GetDB().Delete(&models.Group{}, "id = ?", id).Error
}

// List lists groups with pagination
func (s *GroupService) List(page, pageSize int) ([]models.Group, int64, error) {
	var groups []models.Group
	var total int64

	query := database.GetDB().Model(&models.Group{})
	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// AddUserToGroup adds a user to a group
func (s *GroupService) AddUserToGroup(groupID, userID, createdBy uuid.UUID) error {
	// Verify group exists
	if _, err := s.GetByID(groupID); err != nil {
		return err
	}

	// Verify user exists
	userService := NewUserService()
	if _, err := userService.GetByID(userID); err != nil {
		return err
	}

	userGroup := &models.UserGroup{
		GroupID:   groupID,
		UserID:    userID,
		CreatedBy: createdBy,
	}

	return database.GetDB().Create(userGroup).Error
}

// RemoveUserFromGroup removes a user from a group
func (s *GroupService) RemoveUserFromGroup(groupID, userID uuid.UUID) error {
	return database.GetDB().Delete(&models.UserGroup{}, "group_id = ? AND user_id = ?", groupID, userID).Error
}

// GetGroupUsers gets all users in a group
func (s *GroupService) GetGroupUsers(groupID uuid.UUID) ([]models.User, error) {
	var userGroups []models.UserGroup
	if err := database.GetDB().Preload("User").Where("group_id = ?", groupID).Find(&userGroups).Error; err != nil {
		return nil, err
	}

	users := make([]models.User, len(userGroups))
	for i, ug := range userGroups {
		users[i] = ug.User
	}

	return users, nil
}

// GetUserGroups gets all groups a user belongs to
func (s *GroupService) GetUserGroups(userID uuid.UUID) ([]models.Group, error) {
	var userGroups []models.UserGroup
	if err := database.GetDB().Preload("Group").Where("user_id = ?", userID).Find(&userGroups).Error; err != nil {
		return nil, err
	}

	groups := make([]models.Group, len(userGroups))
	for i, ug := range userGroups {
		groups[i] = ug.Group
	}

	return groups, nil
}
