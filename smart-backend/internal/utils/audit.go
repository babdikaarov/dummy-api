package utils

import (
	"log"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"

	"github.com/google/uuid"
)

// LogAdminAction logs an admin action to the audit log
// This tracks all administrative operations for security and compliance purposes
func LogAdminAction(
	adminID uuid.UUID,
	adminName string,
	action string,           // "create_user", "update_user", etc.
	resourceType string,     // "user", "admin", "contact", etc.
	resourceID string,       // UUID or ID of the resource
	details string,          // JSON string with operation details
	ipAddress string,        // Request IP
	userAgent string,        // Request user agent
	status string,           // "success" or "failed"
	errorMessage string,     // Error message if failed
) {
	auditLog := models.AdminAuditLog{
		ID:           uuid.New(),
		AdminID:      adminID,
		AdminName:    adminName,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Status:       status,
		ErrorMessage: errorMessage,
	}

	if err := db.DB.Create(&auditLog).Error; err != nil {
		log.Printf("Error creating audit log: %v", err)
	}
}
