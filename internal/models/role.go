package models

// Role represents user role in the system
type Role string

const (
	RoleUser    Role = "USER"
	RoleManager Role = "MANAGER"
	RoleAdmin   Role = "ADMIN"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleManager, RoleAdmin:
		return true
	default:
		return false
	}
}

// CanManageUsers checks if the role can manage other users
func (r Role) CanManageUsers() bool {
	return r == RoleManager || r == RoleAdmin
}

// CanManageAll checks if the role has full administrative access
func (r Role) CanManageAll() bool {
	return r == RoleAdmin
}
