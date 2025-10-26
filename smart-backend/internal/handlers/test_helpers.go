package handlers

import (
	"ololo-gate/internal/config"
	"ololo-gate/internal/db"
	"ololo-gate/internal/middleware"
	"ololo-gate/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestApp creates a Fiber app with all routes configured for testing
func SetupTestApp() (*fiber.App, func()) {
	// Setup test config
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key",
			AccessExpiry:  900000000000,      // 15 minutes in nanoseconds
			RefreshExpiry: 2592000000000000,  // 30 days in nanoseconds
		},
		Server: config.ServerConfig{
			Port: "8080",
			Env:  "test",
		},
	}

	// Setup test config for third-party API (use empty URL for tests)
	config.AppConfig.ThirdPartyAPIURL = "http://localhost:3000"

	// Setup test database
	db.DB, _ = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.DB.AutoMigrate(&models.User{}, &models.Admin{}, &models.Contact{}, &models.AdminAuditLog{})

	app := fiber.New()

	// Setup routes exactly as in main.go
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", Register)
	auth.Post("/login", Login)
	auth.Post("/refresh", RefreshToken)
	auth.Get("/check-phone", CheckPhoneAvailability)

	// User management routes (protected - requires Admin JWT authentication)
	users := api.Group("/users", middleware.AdminJWTProtected())
	users.Get("/", GetAllUsers)
	users.Post("/", CreateUser)
	users.Get("/:id", GetUserByID)
	users.Patch("/:id", UpdateUser)
	users.Delete("/:id", DeleteUser)

	// Admin authentication (public)
	adminAuth := api.Group("/admin")
	adminAuth.Post("/login", AdminLogin)

	// Admin user management routes (Admin JWT protected, role-based access control in handlers)
	adminUsers := api.Group("/admin/users", middleware.AdminJWTProtected())
	adminUsers.Get("/", middleware.SuperAdminOnly(), GetAllAdmins)
	adminUsers.Post("/", middleware.SuperAdminOnly(), CreateAdmin)
	adminUsers.Get("/:id", GetAdminByID)
	adminUsers.Patch("/:id", UpdateAdmin)
	adminUsers.Delete("/:id", middleware.SuperAdminOnly(), DeleteAdmin)

	// Gate management routes (User JWT protected - users only, not admins)
	api.Get("/locations", middleware.JWTProtected(), GetLocations)
	api.Get("/locations/:locationId/gates", middleware.JWTProtected(), GetGatesByLocation)
	api.Put("/locations/:gateId/open", middleware.JWTProtected(), OpenGate)
	api.Put("/locations/:gateId/close", middleware.JWTProtected(), CloseGate)

	// Available locations route (Admin JWT protected)
	api.Get("/available-locations", middleware.AdminJWTProtected(), GetAvailableLocations)

	// Contact information routes
	api.Get("/contacts", GetContact)
	api.Patch("/contacts", middleware.AdminJWTProtected(), UpdateContact)

	// Admin audit log routes (Admin JWT protected, super admin only)
	adminAudit := api.Group("/admin/audit-logs", middleware.AdminJWTProtected(), middleware.SuperAdminOnly())
	adminAudit.Get("/", GetAdminAuditLogs)
	adminAudit.Get("/:id", GetAdminAuditLogByID)

	cleanup := func() {
		db.DB.Exec("DELETE FROM users")
		db.DB.Exec("DELETE FROM admins")
		db.DB.Exec("DELETE FROM contacts")
		db.DB.Exec("DELETE FROM admin_audit_logs")
	}

	return app, cleanup
}
