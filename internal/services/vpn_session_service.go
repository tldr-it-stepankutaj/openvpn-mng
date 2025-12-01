package services

import (
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/dto"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

var (
	ErrSessionNotFound = errors.New("vpn session not found")
)

// VpnSessionService provides VPN session services
type VpnSessionService struct{}

// NewVpnSessionService creates a new VPN session service
func NewVpnSessionService() *VpnSessionService {
	return &VpnSessionService{}
}

// List returns a paginated list of VPN sessions with optional filters
func (s *VpnSessionService) List(filter *dto.VpnSessionFilter) ([]models.VpnSession, int64, error) {
	var sessions []models.VpnSession
	var total int64

	query := database.GetDB().Model(&models.VpnSession{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.VpnIP != "" {
		query = query.Where("vpn_ip = ?", filter.VpnIP)
	}
	if filter.IsActive != nil {
		if *filter.IsActive {
			query = query.Where("disconnected_at IS NULL")
		} else {
			query = query.Where("disconnected_at IS NOT NULL")
		}
	}
	if filter.StartDate != nil {
		query = query.Where("connected_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("connected_at <= ?", filter.EndDate)
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

	// Get sessions with preloaded user
	if err := query.Preload("User").Order("connected_at DESC").Offset(offset).Limit(pageSize).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// GetByID returns a VPN session by ID
func (s *VpnSessionService) GetByID(id uuid.UUID) (*models.VpnSession, error) {
	var session models.VpnSession
	if err := database.GetDB().Preload("User").First(&session, "id = ?", id).Error; err != nil {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

// Create creates a new VPN session
func (s *VpnSessionService) Create(req *dto.CreateVpnSessionRequest) (*models.VpnSession, error) {
	session := &models.VpnSession{
		UserID:      req.UserID,
		VpnIP:       req.VpnIP,
		ClientIP:    req.ClientIP,
		ConnectedAt: req.ConnectedAt,
	}

	if err := database.GetDB().Create(session).Error; err != nil {
		return nil, err
	}

	// Reload with user
	return s.GetByID(session.ID)
}

// Disconnect updates a VPN session with disconnect info
func (s *VpnSessionService) Disconnect(id uuid.UUID, req *dto.UpdateVpnSessionRequest) (*models.VpnSession, error) {
	session, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update disconnect info
	session.DisconnectedAt = &req.DisconnectedAt
	session.BytesReceived = req.BytesReceived
	session.BytesSent = req.BytesSent
	session.DisconnectReason = req.DisconnectReason

	if err := database.GetDB().Save(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

// GetActiveSessions returns all active (not disconnected) sessions
func (s *VpnSessionService) GetActiveSessions() ([]models.VpnSession, error) {
	var sessions []models.VpnSession
	if err := database.GetDB().Preload("User").Where("disconnected_at IS NULL").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetActiveSessionByUserID returns the active session for a user (if any)
func (s *VpnSessionService) GetActiveSessionByUserID(userID uuid.UUID) (*models.VpnSession, error) {
	var session models.VpnSession
	if err := database.GetDB().Preload("User").Where("user_id = ? AND disconnected_at IS NULL", userID).First(&session).Error; err != nil {
		return nil, ErrSessionNotFound
	}
	return &session, nil
}

// GetUsageStats returns aggregated usage statistics
func (s *VpnSessionService) GetUsageStats() (*dto.VpnUsageStats, error) {
	var stats dto.VpnUsageStats

	// Total sessions
	if err := database.GetDB().Model(&models.VpnSession{}).Count(&stats.TotalSessions).Error; err != nil {
		return nil, err
	}

	// Active sessions
	if err := database.GetDB().Model(&models.VpnSession{}).Where("disconnected_at IS NULL").Count(&stats.ActiveSessions).Error; err != nil {
		return nil, err
	}

	// Total bytes
	type ByteStats struct {
		TotalReceived int64
		TotalSent     int64
	}
	var byteStats ByteStats
	if err := database.GetDB().Model(&models.VpnSession{}).
		Select("COALESCE(SUM(bytes_received), 0) as total_received, COALESCE(SUM(bytes_sent), 0) as total_sent").
		Scan(&byteStats).Error; err != nil {
		return nil, err
	}
	stats.TotalBytesReceived = byteStats.TotalReceived
	stats.TotalBytesSent = byteStats.TotalSent
	stats.TotalBytes = byteStats.TotalReceived + byteStats.TotalSent

	return &stats, nil
}

// GetUserUsageStats returns usage statistics per user
func (s *VpnSessionService) GetUserUsageStats(page, pageSize int) ([]dto.UserVpnUsageResponse, int64, error) {
	var total int64

	// Count unique users with sessions
	if err := database.GetDB().Model(&models.VpnSession{}).
		Select("COUNT(DISTINCT user_id)").
		Scan(&total).Error; err != nil {
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

	// Get aggregated stats per user
	type UserStats struct {
		UserID        uuid.UUID
		Username      string
		TotalSessions int64
		TotalReceived int64
		TotalSent     int64
		LastConnected *string
	}
	var userStats []UserStats

	if err := database.GetDB().Table("vpn_sessions").
		Select(`
			vpn_sessions.user_id,
			users.username,
			COUNT(*) as total_sessions,
			COALESCE(SUM(vpn_sessions.bytes_received), 0) as total_received,
			COALESCE(SUM(vpn_sessions.bytes_sent), 0) as total_sent,
			MAX(vpn_sessions.connected_at) as last_connected
		`).
		Joins("JOIN users ON users.id = vpn_sessions.user_id").
		Group("vpn_sessions.user_id, users.username").
		Order("total_sessions DESC").
		Offset(offset).
		Limit(pageSize).
		Scan(&userStats).Error; err != nil {
		return nil, 0, err
	}

	// Convert to response DTOs
	responses := make([]dto.UserVpnUsageResponse, len(userStats))
	for i, us := range userStats {
		responses[i] = dto.UserVpnUsageResponse{
			UserID:             us.UserID,
			Username:           us.Username,
			TotalSessions:      us.TotalSessions,
			TotalBytesReceived: us.TotalReceived,
			TotalBytesSent:     us.TotalSent,
			TotalBytes:         us.TotalReceived + us.TotalSent,
		}
	}

	return responses, total, nil
}

// VpnTrafficStatsService provides VPN traffic stats services
type VpnTrafficStatsService struct{}

// NewVpnTrafficStatsService creates a new VPN traffic stats service
func NewVpnTrafficStatsService() *VpnTrafficStatsService {
	return &VpnTrafficStatsService{}
}

// List returns a paginated list of traffic stats with optional filters
func (s *VpnTrafficStatsService) List(filter *dto.VpnTrafficStatsFilter) ([]models.VpnTrafficStats, int64, error) {
	var stats []models.VpnTrafficStats
	var total int64

	query := database.GetDB().Model(&models.VpnTrafficStats{})

	// Apply filters
	if filter.SessionID != nil {
		query = query.Where("session_id = ?", filter.SessionID)
	}
	if filter.StartDate != nil {
		query = query.Where("timestamp >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("timestamp <= ?", filter.EndDate)
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

	// Get stats
	if err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&stats).Error; err != nil {
		return nil, 0, err
	}

	return stats, total, nil
}

// Create creates a new traffic stats entry
func (s *VpnTrafficStatsService) Create(req *dto.CreateVpnTrafficStatsRequest) (*models.VpnTrafficStats, error) {
	stats := &models.VpnTrafficStats{
		SessionID:          req.SessionID,
		Timestamp:          req.Timestamp,
		BytesReceivedDelta: req.BytesReceivedDelta,
		BytesSentDelta:     req.BytesSentDelta,
	}

	if err := database.GetDB().Create(stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetBySessionID returns all traffic stats for a session
func (s *VpnTrafficStatsService) GetBySessionID(sessionID uuid.UUID, page, pageSize int) ([]models.VpnTrafficStats, int64, error) {
	var stats []models.VpnTrafficStats
	var total int64

	query := database.GetDB().Model(&models.VpnTrafficStats{}).Where("session_id = ?", sessionID)

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

	if err := query.Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&stats).Error; err != nil {
		return nil, 0, err
	}

	return stats, total, nil
}

// CalculateTotalPages calculates total pages for pagination
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		pageSize = 20
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}
