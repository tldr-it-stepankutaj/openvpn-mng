package middleware

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
)

// AuditLogger provides methods for logging audit events
type AuditLogger struct{}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// Log creates an audit log entry
func (al *AuditLogger) Log(c *gin.Context, action models.AuditAction, entityType string, entityID *uuid.UUID, oldValues, newValues interface{}, details string) error {
	authUser := GetAuthUser(c)
	if authUser == nil {
		return nil // Skip audit if no user is authenticated
	}

	userID, err := uuid.Parse(authUser.ID)
	if err != nil {
		return err
	}

	var oldValuesStr, newValuesStr string
	if oldValues != nil {
		bytes, _ := json.Marshal(oldValues)
		oldValuesStr = string(bytes)
	}
	if newValues != nil {
		bytes, _ := json.Marshal(newValues)
		newValuesStr = string(bytes)
	}

	auditLog := &models.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldValuesStr,
		NewValues:  newValuesStr,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Details:    details,
	}

	return database.GetDB().Create(auditLog).Error
}

// LogCreate logs a create action
func (al *AuditLogger) LogCreate(c *gin.Context, entityType string, entityID uuid.UUID, newValues interface{}) error {
	return al.Log(c, models.AuditActionCreate, entityType, &entityID, nil, newValues, "")
}

// LogUpdate logs an update action
func (al *AuditLogger) LogUpdate(c *gin.Context, entityType string, entityID uuid.UUID, oldValues, newValues interface{}) error {
	return al.Log(c, models.AuditActionUpdate, entityType, &entityID, oldValues, newValues, "")
}

// LogDelete logs a delete action
func (al *AuditLogger) LogDelete(c *gin.Context, entityType string, entityID uuid.UUID, oldValues interface{}) error {
	return al.Log(c, models.AuditActionDelete, entityType, &entityID, oldValues, nil, "")
}

// LogLogin logs a login action
func (al *AuditLogger) LogLogin(c *gin.Context, userID uuid.UUID, details string) error {
	auditLog := &models.AuditLog{
		UserID:     userID,
		Action:     models.AuditActionLogin,
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Details:    details,
	}
	return database.GetDB().Create(auditLog).Error
}

// LogLogout logs a logout action
func (al *AuditLogger) LogLogout(c *gin.Context, userID uuid.UUID) error {
	auditLog := &models.AuditLog{
		UserID:     userID,
		Action:     models.AuditActionLogout,
		EntityType: "user",
		EntityID:   &userID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}
	return database.GetDB().Create(auditLog).Error
}
