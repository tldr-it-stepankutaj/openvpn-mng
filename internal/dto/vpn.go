package dto

// NextVPNIPResponse represents the response for next available VPN IP
type NextVPNIPResponse struct {
	IP string `json:"ip"`
}

// ValidateVPNIPRequest represents a request to validate a VPN IP
type ValidateVPNIPRequest struct {
	IP            string `json:"ip" binding:"required"`
	ExcludeUserID string `json:"exclude_user_id,omitempty"`
}

// ValidateVPNIPResponse represents the response for IP validation
type ValidateVPNIPResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// UsedVPNIPsResponse represents the response for used VPN IPs
type UsedVPNIPsResponse struct {
	IPs []string `json:"ips"`
}
