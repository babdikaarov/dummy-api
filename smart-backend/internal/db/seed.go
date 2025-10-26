package db

import (
	"log"
	"ololo-gate/internal/config"
	"ololo-gate/internal/models"

	"github.com/google/uuid"
)

// CreateInitialAdmin creates the initial super admin if it doesn't exist
func CreateInitialAdmin() {
	adminConfig := config.AppConfig.InitAdmin

	// Parse UUID from config
	adminUUID, err := uuid.Parse(adminConfig.UUID)
	if err != nil {
		log.Fatalf("Invalid INIT_ADMIN_UUID format: %v", err)
	}

	// Check if admin with this UUID already exists
	var existingAdmin models.Admin
	result := DB.Where("id = ?", adminUUID).First(&existingAdmin)

	if result.Error == nil {
		// Admin already exists
		log.Printf("ℹ️  Initial admin already exists (ID: %s, Username: %s)", adminUUID, existingAdmin.Username)
		return
	}

	// Create initial super admin
	initialAdmin := models.Admin{
		ID:       adminUUID,
		Username: adminConfig.Username,
		Password: adminConfig.Password, // Will be hashed by BeforeCreate hook
		Role:     models.RoleSuper,
	}

	if err := DB.Create(&initialAdmin).Error; err != nil {
		log.Fatalf("Failed to create initial admin: %v", err)
	}

	log.Printf("✅ Initial super admin created successfully (Username: %s)", adminConfig.Username)
	log.Printf("⚠️  Please change the default admin password in production!")
}
