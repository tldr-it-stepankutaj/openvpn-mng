package services

import (
	"encoding/binary"
	"errors"
	"net"
	"sort"

	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

var (
	ErrVPNNetworkNotConfigured = errors.New("VPN network not configured")
	ErrInvalidVPNNetwork       = errors.New("invalid VPN network CIDR")
	ErrNoAvailableIP           = errors.New("no available IP addresses in VPN network")
	ErrIPOutOfRange            = errors.New("IP address is outside VPN network range")
	ErrIPAlreadyUsed           = errors.New("IP address is already in use")
	ErrIPReservedForServer     = errors.New("IP address is reserved for VPN server")
)

// VPNIPService provides VPN IP allocation services
type VPNIPService struct {
	config *config.VPNConfig
}

// NewVPNIPService creates a new VPN IP service
func NewVPNIPService(cfg *config.VPNConfig) *VPNIPService {
	return &VPNIPService{config: cfg}
}

// GetNextAvailableIP returns the next available IP in the VPN network
func (s *VPNIPService) GetNextAvailableIP() (string, error) {
	if s.config.Network == "" {
		return "", ErrVPNNetworkNotConfigured
	}

	_, ipNet, err := net.ParseCIDR(s.config.Network)
	if err != nil {
		return "", ErrInvalidVPNNetwork
	}

	// Get all used IPs from database
	usedIPs, err := s.getUsedIPs()
	if err != nil {
		return "", err
	}

	// Add server IP to used IPs
	if s.config.ServerIP != "" {
		usedIPs[s.config.ServerIP] = true
	}

	// Find first available IP
	ip := ipNet.IP.Mask(ipNet.Mask)
	for {
		ip = nextIP(ip)

		// Check if we're still in the network
		if !ipNet.Contains(ip) {
			return "", ErrNoAvailableIP
		}

		ipStr := ip.String()

		// Skip network address (first IP) and broadcast (last IP)
		if isNetworkOrBroadcast(ip, ipNet) {
			continue
		}

		// Skip if already used
		if usedIPs[ipStr] {
			continue
		}

		return ipStr, nil
	}
}

// ValidateIP validates that an IP is within the VPN network and not already used
func (s *VPNIPService) ValidateIP(ipStr string, excludeUserID ...string) error {
	if s.config.Network == "" {
		return ErrVPNNetworkNotConfigured
	}

	_, ipNet, err := net.ParseCIDR(s.config.Network)
	if err != nil {
		return ErrInvalidVPNNetwork
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errors.New("invalid IP address format")
	}

	// Check if IP is in network range
	if !ipNet.Contains(ip) {
		return ErrIPOutOfRange
	}

	// Check if it's the server IP
	if s.config.ServerIP != "" && ipStr == s.config.ServerIP {
		return ErrIPReservedForServer
	}

	// Check if IP is already used by another user
	var count int64
	query := database.GetDB().Model(&models.User{}).Where("vpn_ip = ?", ipStr)

	// Exclude specific user (for updates)
	if len(excludeUserID) > 0 && excludeUserID[0] != "" {
		query = query.Where("id != ?", excludeUserID[0])
	}

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return ErrIPAlreadyUsed
	}

	return nil
}

// GetNetworkInfo returns information about the VPN network
func (s *VPNIPService) GetNetworkInfo() (*VPNNetworkInfo, error) {
	if s.config.Network == "" {
		return nil, ErrVPNNetworkNotConfigured
	}

	_, ipNet, err := net.ParseCIDR(s.config.Network)
	if err != nil {
		return nil, ErrInvalidVPNNetwork
	}

	// Calculate total usable IPs (excluding network and broadcast)
	ones, bits := ipNet.Mask.Size()
	totalIPs := (1 << (bits - ones)) - 2 // -2 for network and broadcast
	if totalIPs < 0 {
		totalIPs = 0
	}

	// Get used count
	usedIPs, err := s.getUsedIPs()
	if err != nil {
		return nil, err
	}

	usedCount := len(usedIPs)
	if s.config.ServerIP != "" {
		usedCount++ // Include server IP in used count
	}

	return &VPNNetworkInfo{
		Network:      s.config.Network,
		ServerIP:     s.config.ServerIP,
		TotalIPs:     totalIPs,
		UsedIPs:      usedCount,
		AvailableIPs: totalIPs - usedCount,
	}, nil
}

// VPNNetworkInfo contains information about the VPN network
type VPNNetworkInfo struct {
	Network      string `json:"network"`
	ServerIP     string `json:"server_ip"`
	TotalIPs     int    `json:"total_ips"`
	UsedIPs      int    `json:"used_ips"`
	AvailableIPs int    `json:"available_ips"`
}

// getUsedIPs returns a map of all VPN IPs currently in use
func (s *VPNIPService) getUsedIPs() (map[string]bool, error) {
	var users []models.User
	if err := database.GetDB().Select("vpn_ip").Where("vpn_ip IS NOT NULL AND vpn_ip != ''").Find(&users).Error; err != nil {
		return nil, err
	}

	usedIPs := make(map[string]bool)
	for _, user := range users {
		if user.VpnIP != "" {
			usedIPs[user.VpnIP] = true
		}
	}

	return usedIPs, nil
}

// GetUsedIPs returns a sorted list of all used VPN IPs
func (s *VPNIPService) GetUsedIPs() ([]string, error) {
	usedIPs, err := s.getUsedIPs()
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(usedIPs))
	for ip := range usedIPs {
		ips = append(ips, ip)
	}

	// Sort IPs numerically
	sort.Slice(ips, func(i, j int) bool {
		return ipToInt(net.ParseIP(ips[i])) < ipToInt(net.ParseIP(ips[j]))
	})

	return ips, nil
}

// nextIP returns the next IP address
func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}

	return next
}

// isNetworkOrBroadcast checks if IP is network address or broadcast
func isNetworkOrBroadcast(ip net.IP, ipNet *net.IPNet) bool {
	// Get the IP in 4-byte format
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	// Network address check
	network := ipNet.IP.Mask(ipNet.Mask)
	if ip4.Equal(network) {
		return true
	}

	// Broadcast address check
	broadcast := make(net.IP, len(ip4))
	for i := 0; i < len(ip4); i++ {
		broadcast[i] = ip4[i] | ^ipNet.Mask[i]
	}

	return ip4.Equal(broadcast)
}

// ipToInt converts an IP to uint32 for sorting
func ipToInt(ip net.IP) uint32 {
	if ip == nil {
		return 0
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip4)
}
