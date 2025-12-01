package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

var (
	ErrAuditLogNotFound = errors.New("audit log not found")
)

// AuditService provides audit log services
type AuditService struct{}

// NewAuditService creates a new audit service
func NewAuditService() *AuditService {
	return &AuditService{}
}

// List returns a paginated list of audit logs with optional filters
func (s *AuditService) List(filter *dto.AuditLogFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.GetDB().Model(&models.AuditLog{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Action != nil {
		query = query.Where("action = ?", filter.Action)
	}
	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}
	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", filter.EntityID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Get logs with preloaded user
	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID returns an audit log by ID
func (s *AuditService) GetByID(id uuid.UUID) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := database.GetDB().Preload("User").First(&log, "id = ?", id).Error; err != nil {
		return nil, ErrAuditLogNotFound
	}
	return &log, nil
}

// GetByEntityID returns audit logs for a specific entity
func (s *AuditService) GetByEntityID(entityType string, entityID uuid.UUID, page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.GetDB().Model(&models.AuditLog{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUserID returns audit logs for a specific user
func (s *AuditService) GetByUserID(userID uuid.UUID, page, pageSize int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := database.GetDB().Model(&models.AuditLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	if err := query.Preload("User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetActions returns all available audit actions
func (s *AuditService) GetActions() []models.AuditAction {
	return []models.AuditAction{
		models.AuditActionCreate,
		models.AuditActionRead,
		models.AuditActionUpdate,
		models.AuditActionDelete,
		models.AuditActionLogin,
		models.AuditActionLogout,
	}
}

// GetEntityTypes returns all entity types that have been logged
func (s *AuditService) GetEntityTypes() ([]string, error) {
	var types []string
	if err := database.GetDB().Model(&models.AuditLog{}).
		Distinct("entity_type").
		Pluck("entity_type", &types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// GetStatsByAction returns count of audit logs grouped by action
func (s *AuditService) GetStatsByAction() (map[models.AuditAction]int64, error) {
	type ActionCount struct {
		Action models.AuditAction
		Count  int64
	}
	var results []ActionCount

	if err := database.GetDB().Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[models.AuditAction]int64)
	for _, r := range results {
		stats[r.Action] = r.Count
	}
	return stats, nil
}

// GetStatsByEntityType returns count of audit logs grouped by entity type
func (s *AuditService) GetStatsByEntityType() (map[string]int64, error) {
	type EntityCount struct {
		EntityType string
		Count      int64
	}
	var results []EntityCount

	if err := database.GetDB().Model(&models.AuditLog{}).
		Select("entity_type, COUNT(*) as count").
		Group("entity_type").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.EntityType] = r.Count
	}
	return stats, nil
}
