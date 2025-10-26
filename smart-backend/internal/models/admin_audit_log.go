package models

import (
	"time"

	"github.com/google/uuid"
)

// AdminAuditLog represents an audit log entry for admin actions
type AdminAuditLog struct {
	ID           uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	AdminID      uuid.UUID `gorm:"type:char(36);index" json:"admin_id"`          // Who performed the action
	AdminName    string    `gorm:"index" json:"admin_name"`                      // Admin username for quick access (denormalized)
	Action       string    `gorm:"index" json:"action"`                          // "create_user", "update_user", "delete_user", "create_admin", "delete_admin", "update_contact", etc.
	ResourceType string    `gorm:"index" json:"resource_type"`                   // "user", "admin", "contact", etc.
	ResourceID   string    `gorm:"index" json:"resource_id"`                     // UUID or ID of affected resource
	Details      string    `gorm:"type:text" json:"details"`                     // JSON with request details (what was changed)
	IPAddress    string    `json:"ip_address"`                                    // Request IP address
	UserAgent    string    `gorm:"type:text" json:"user_agent"`                  // Request user agent
	Status       string    `json:"status"`                                        // "success" or "failed"
	ErrorMessage string    `gorm:"type:text" json:"error_message"`               // Error message if failed
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for the AdminAuditLog model
func (AdminAuditLog) TableName() string {
	return "admin_audit_logs"
}
