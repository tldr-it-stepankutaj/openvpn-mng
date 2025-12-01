package services

import (
	"errors"
	"net"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"gorm.io/gorm"
)

var (
	ErrNetworkExists   = errors.New("network already exists")
	ErrNetworkNotFound = errors.New("network not found")
	ErrInvalidCIDR     = errors.New("invalid CIDR format")
)

// NetworkService provides network management services
type NetworkService struct{}

// NewNetworkService creates a new network service
func NewNetworkService() *NetworkService {
	return &NetworkService{}
}

// validateCIDR validates that the given string is a valid CIDR notation or IP address
func (s *NetworkService) validateCIDR(cidr string) error {
	// Try parsing as CIDR
	_, _, err := net.ParseCIDR(cidr)
	if err == nil {
		return nil
	}

	// Try parsing as single IP and convert to CIDR
	ip := net.ParseIP(cidr)
	if ip != nil {
		return nil // Valid single IP
	}

	return ErrInvalidCIDR
}

// normalizeCIDR normalizes the CIDR (adds /32 or /128 if it's a single IP)
func (s *NetworkService) normalizeCIDR(cidr string) string {
	// Check if it's already a valid CIDR
	_, _, err := net.ParseCIDR(cidr)
	if err == nil {
		return cidr
	}

	// Try to parse as single IP
	ip := net.ParseIP(cidr)
	if ip != nil {
		if ip.To4() != nil {
			return cidr + "/32"
		}
		return cidr + "/128"
	}

	return cidr
}

// Create creates a new network
func (s *NetworkService) Create(req *dto.CreateNetworkRequest, createdBy uuid.UUID) (*models.Network, error) {
	// Validate CIDR
	if err := s.validateCIDR(req.CIDR); err != nil {
		return nil, err
	}

	// Normalize CIDR
	normalizedCIDR := s.normalizeCIDR(req.CIDR)

	// Check if network with same name already exists
	var existing models.Network
	if err := database.GetDB().Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, ErrNetworkExists
	}

	network := &models.Network{
		Name:        req.Name,
		CIDR:        normalizedCIDR,
		Description: req.Description,
		CreatedBy:   createdBy,
	}

	if err := database.GetDB().Create(network).Error; err != nil {
		return nil, err
	}

	return network, nil
}

// GetByID gets a network by ID
func (s *NetworkService) GetByID(id uuid.UUID) (*models.Network, error) {
	var network models.Network
	if err := database.GetDB().Preload("Groups").Preload("Creator").Preload("Updater").First(&network, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNetworkNotFound
		}
		return nil, err
	}
	return &network, nil
}

// Update updates a network
func (s *NetworkService) Update(id uuid.UUID, req *dto.UpdateNetworkRequest, updatedBy uuid.UUID) (*models.Network, error) {
	network, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.CIDR != "" {
		if err := s.validateCIDR(req.CIDR); err != nil {
			return nil, err
		}
		updates["cidr"] = s.normalizeCIDR(req.CIDR)
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	updates["updated_by"] = updatedBy

	if err := database.GetDB().Model(network).Updates(updates).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

// Delete soft deletes a network
func (s *NetworkService) Delete(id uuid.UUID) error {
	// First remove all group associations
	if err := database.GetDB().Where("network_id = ?", id).Delete(&models.NetworkGroup{}).Error; err != nil {
		return err
	}
	return database.GetDB().Delete(&models.Network{}, "id = ?", id).Error
}

// List lists networks with pagination
func (s *NetworkService) List(page, pageSize int) ([]models.Network, int64, error) {
	var networks []models.Network
	var total int64

	query := database.GetDB().Model(&models.Network{}).Preload("Groups")

	// Count total
	query.Count(&total)

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&networks).Error; err != nil {
		return nil, 0, err
	}

	return networks, total, nil
}

// AddGroupToNetwork adds a group to a network
func (s *NetworkService) AddGroupToNetwork(networkID, groupID, createdBy uuid.UUID) error {
	// Check if network exists
	if _, err := s.GetByID(networkID); err != nil {
		return err
	}

	// Check if group exists
	groupService := NewGroupService()
	if _, err := groupService.GetByID(groupID); err != nil {
		return err
	}

	// Check if association already exists
	var existing models.NetworkGroup
	if err := database.GetDB().Where("network_id = ? AND group_id = ?", networkID, groupID).First(&existing).Error; err == nil {
		return errors.New("group already associated with this network")
	}

	networkGroup := &models.NetworkGroup{
		NetworkID: networkID,
		GroupID:   groupID,
		CreatedBy: createdBy,
	}

	return database.GetDB().Create(networkGroup).Error
}

// RemoveGroupFromNetwork removes a group from a network
func (s *NetworkService) RemoveGroupFromNetwork(networkID, groupID uuid.UUID) error {
	return database.GetDB().Where("network_id = ? AND group_id = ?", networkID, groupID).Delete(&models.NetworkGroup{}).Error
}

// GetNetworkGroups gets all groups for a network
func (s *NetworkService) GetNetworkGroups(networkID uuid.UUID) ([]models.Group, error) {
	var groups []models.Group

	err := database.GetDB().
		Joins("JOIN network_groups ON network_groups.group_id = groups.id").
		Where("network_groups.network_id = ?", networkID).
		Find(&groups).Error

	return groups, err
}

// GetGroupNetworks gets all networks for a group
func (s *NetworkService) GetGroupNetworks(groupID uuid.UUID) ([]models.Network, error) {
	var networks []models.Network

	err := database.GetDB().
		Joins("JOIN network_groups ON network_groups.network_id = networks.id").
		Where("network_groups.group_id = ?", groupID).
		Find(&networks).Error

	return networks, err
}
