package main

import (
	"fmt"
	"log"
	"ololo-gate/internal/config"
	"ololo-gate/internal/db"
	"ololo-gate/internal/handlers"
	"ololo-gate/internal/middleware"
	"ololo-gate/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	fiberSwagger "github.com/swaggo/fiber-swagger"
	_ "ololo-gate/docs" // Import generated docs
)

// serverStartTime tracks when the server started for uptime calculation
var serverStartTime time.Time

// @title Ololo Gate API
// @version 1.0
// @description Secure phone-based authentication backend for Ololo Gate management system with dual authentication (users & admins) and role-based access control.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@ololo-gate.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Record server start time
	serverStartTime = time.Now()

	// Load configuration
	config.LoadConfig()

	// Connect to database
	db.Connect()

	// Auto-migrate database models
	db.AutoMigrate(&models.User{}, &models.Admin{}, &models.Contact{}, &models.AdminAuditLog{})

	// Create initial super admin if not exists
	db.CreateInitialAdmin()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Ololo Gate API v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New()) // Recover from panics
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))

	// CORS configuration - handle wildcard origins securely
	corsConfig := cors.Config{
		AllowOrigins:     config.AppConfig.CORS.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400, // 24 hours preflight cache
		AllowCredentials: config.AppConfig.CORS.AllowedOrigins != "*", // Only allow credentials if not using wildcard
	}
	app.Use(cors.New(corsConfig))

	// Routes
	setupRoutes(app)

	// Start server
	port := ":" + config.AppConfig.Server.Port
	log.Printf("🚀 Ololo Gate API server starting on port %s", config.AppConfig.Server.Port)
	log.Fatal(app.Listen(port))
}

func setupRoutes(app *fiber.App) {
	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Health check endpoint
	app.Get("/", healthCheck)

	// API v1 routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", handlers.Register)                    // POST /api/v1/auth/register - Register new user
	auth.Post("/login", handlers.Login)                          // POST /api/v1/auth/login - Login user
	auth.Post("/refresh", handlers.RefreshToken)                 // POST /api/v1/auth/refresh - Refresh access token
	auth.Get("/check-phone", handlers.CheckPhoneAvailability)    // GET /api/v1/auth/check-phone - Check if phone number is available

	// User management routes (protected - requires Admin JWT authentication)
	users := api.Group("/users", middleware.AdminJWTProtected())
	users.Get("/", handlers.GetAllUsers)        // GET /api/v1/users - Get all users (admins only)
	users.Post("/", handlers.CreateUser)        // POST /api/v1/users - Create new user with locations/gates (admins only)
	users.Get("/:id", handlers.GetUserByID)     // GET /api/v1/users/:id - Get user by ID (admins only)
	users.Patch("/:id", handlers.UpdateUser)    // PATCH /api/v1/users/:id - Update user password and locations/gates (admins only)
	users.Delete("/:id", handlers.DeleteUser)   // DELETE /api/v1/users/:id - Delete user (admins only)

	// Admin authentication (public)
	adminAuth := api.Group("/admin")
	adminAuth.Post("/login", handlers.AdminLogin) // POST /api/v1/admin/login - Admin login

	// Admin user management routes (Admin JWT protected, role-based access control in handlers)
	adminUsers := api.Group("/admin/users", middleware.AdminJWTProtected())
	adminUsers.Get("/", middleware.SuperAdminOnly(), handlers.GetAllAdmins)           // GET /api/v1/admin/users - Get all admin accounts (super admin only)
	adminUsers.Post("/", middleware.SuperAdminOnly(), handlers.CreateAdmin)           // POST /api/v1/admin/users - Create new admin account (super admin only)
	adminUsers.Get("/:id", handlers.GetAdminByID)                                      // GET /api/v1/admin/users/:id - Get admin by ID (super/regular with self-access)
	adminUsers.Patch("/:id", handlers.UpdateAdmin)                                    // PATCH /api/v1/admin/users/:id - Update admin (super/regular with field-level access)
	adminUsers.Delete("/:id", middleware.SuperAdminOnly(), handlers.DeleteAdmin)      // DELETE /api/v1/admin/users/:id - Delete admin (super admin only)

	// Gate management routes (User JWT protected - users only, not admins)
	api.Get("/locations", middleware.JWTProtected(), handlers.GetLocations)                           // GET /api/v1/locations - Get all locations accessible to user
	api.Get("/locations/:locationId/gates", middleware.JWTProtected(), handlers.GetGatesByLocation)  // GET /api/v1/locations/:locationId/gates - Get gates for location accessible to user
	api.Put("/locations/:gateId/open", middleware.JWTProtected(), handlers.OpenGate)                 // PUT /api/v1/locations/:gateId/open - Open a gate
	api.Put("/locations/:gateId/close", middleware.JWTProtected(), handlers.CloseGate)               // PUT /api/v1/locations/:gateId/close - Close a gate

	// Available locations route (Admin JWT protected - for admin panel to view all available locations)
	api.Get("/available-locations", middleware.AdminJWTProtected(), handlers.GetAvailableLocations)  // GET /api/v1/available-locations - Get all locations in system (admin only)

	// Contact information routes
	api.Get("/contacts", handlers.GetContact)                                  // GET /api/v1/contacts - Get contact information (public)
	api.Patch("/contacts", middleware.AdminJWTProtected(), handlers.UpdateContact) // PATCH /api/v1/contacts - Update contact information (admin only)
}

// healthCheck godoc
// @Summary Health check endpoint
// @Description Check if the API server is running and retrieve detailed health information including status, timestamp, uptime, and environment
// @Tags Health
// @Produce json
// @Success 200 {object} handlers.HealthCheckResponse "Health check successful"
// @Router / [get]
func healthCheck(c *fiber.Ctx) error {
	// Calculate uptime
	uptime := time.Since(serverStartTime)

	// Format uptime as human-readable string
	uptimeStr := formatDuration(uptime)

	// Get current timestamp
	currentTime := time.Now()

	return c.JSON(handlers.HealthCheckResponse{
		Success:     true,
		Message:     "Ololo Gate API is running",
		Status:      "healthy",
		Timestamp:   currentTime.Format(time.RFC3339),
		Uptime:      uptimeStr,
		Environment: config.AppConfig.Server.Env,
		Version:     "1.0.0",
	})
}

// formatDuration converts a time.Duration to a human-readable format
// Example: 1h30m45s, 5m10s, 30s
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, secs)
	} else {
		return fmt.Sprintf("%ds", secs)
	}
}
