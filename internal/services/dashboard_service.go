package services

import (
	"time"

	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// DashboardStats represents dashboard statistics for admin
type DashboardStats struct {
	TotalUsers      int64               `json:"total_users"`
	ActiveUsers     int64               `json:"active_users"`
	ConnectedUsers  int64               `json:"connected_users"`
	TotalGroups     int64               `json:"total_groups"`
	TotalNetworks   int64               `json:"total_networks"`
	TotalSessions   int64               `json:"total_sessions"`
	TrafficStats    []DailyTrafficStats `json:"traffic_stats"`
	RecentAuditLogs []models.AuditLog   `json:"recent_audit_logs"`
}

// DailyTrafficStats represents daily traffic statistics
type DailyTrafficStats struct {
	Date          string `json:"date"`
	BytesReceived int64  `json:"bytes_received"`
	BytesSent     int64  `json:"bytes_sent"`
}

// DashboardService provides dashboard statistics
type DashboardService struct{}

// NewDashboardService creates a new dashboard service
func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// GetAdminStats returns statistics for admin dashboard
func (s *DashboardService) GetAdminStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Total users
	if err := database.GetDB().Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	// Active users (is_active = true)
	if err := database.GetDB().Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	// Connected users (active VPN sessions)
	if err := database.GetDB().Model(&models.VpnSession{}).Where("disconnected_at IS NULL").Count(&stats.ConnectedUsers).Error; err != nil {
		return nil, err
	}

	// Total groups
	if err := database.GetDB().Model(&models.Group{}).Count(&stats.TotalGroups).Error; err != nil {
		return nil, err
	}

	// Total networks
	if err := database.GetDB().Model(&models.Network{}).Count(&stats.TotalNetworks).Error; err != nil {
		return nil, err
	}

	// Total sessions
	if err := database.GetDB().Model(&models.VpnSession{}).Count(&stats.TotalSessions).Error; err != nil {
		return nil, err
	}

	// Traffic stats for last 30 days
	trafficStats, err := s.getTrafficStats(30)
	if err != nil {
		return nil, err
	}
	stats.TrafficStats = trafficStats

	// Recent audit logs (last 10)
	var auditLogs []models.AuditLog
	if err := database.GetDB().Preload("User").Order("created_at DESC").Limit(10).Find(&auditLogs).Error; err != nil {
		return nil, err
	}
	stats.RecentAuditLogs = auditLogs

	return stats, nil
}

// getTrafficStats returns daily traffic stats for the last N days
func (s *DashboardService) getTrafficStats(days int) ([]DailyTrafficStats, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	type DayStats struct {
		Day           string
		BytesReceived int64
		BytesSent     int64
	}
	var results []DayStats

	// Query to get daily aggregated traffic
	dbType := database.GetDB().Dialector.Name()
	var dateFormat string
	if dbType == "postgres" {
		dateFormat = "TO_CHAR(disconnected_at, 'YYYY-MM-DD')"
	} else {
		// MySQL
		dateFormat = "DATE_FORMAT(disconnected_at, '%Y-%m-%d')"
	}

	query := database.GetDB().Table("vpn_sessions").
		Select(dateFormat+" as day, COALESCE(SUM(bytes_received), 0) as bytes_received, COALESCE(SUM(bytes_sent), 0) as bytes_sent").
		Where("disconnected_at IS NOT NULL AND disconnected_at >= ?", startDate).
		Group("day").
		Order("day ASC")

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	// Convert to response format and fill in missing days with zeros
	statsMap := make(map[string]DailyTrafficStats)
	for _, r := range results {
		statsMap[r.Day] = DailyTrafficStats{
			Date:          r.Day,
			BytesReceived: r.BytesReceived,
			BytesSent:     r.BytesSent,
		}
	}

	// Generate all days in range
	stats := make([]DailyTrafficStats, 0, days)
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if s, ok := statsMap[date]; ok {
			stats = append(stats, s)
		} else {
			stats = append(stats, DailyTrafficStats{
				Date:          date,
				BytesReceived: 0,
				BytesSent:     0,
			})
		}
	}

	return stats, nil
}

// GetManagerSubordinateCount returns the count of subordinates for a manager
func (s *DashboardService) GetManagerSubordinateCount(managerID string) (int64, error) {
	var count int64
	if err := database.GetDB().Model(&models.User{}).Where("manager_id = ?", managerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
