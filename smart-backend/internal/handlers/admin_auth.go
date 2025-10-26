package handlers

import (
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// AdminLoginRequest defines the structure for admin login requests
// @name AdminLoginRequest
type AdminLoginRequest struct {
	Username string `json:"username" validate:"required" example:"admin"`
	Password string `json:"password" validate:"required" example:"admin"`
}

// AdminLogin godoc
// @Summary Admin login
// @Description Authenticate admin with username and password, returns permanent access token (no expiry)
// @Tags Admin Authentication
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin credentials"
// @Success 200 {object} AdminLoginResponse "Login successful with permanent token"
// @Failure 400 {object} APIResponse "Invalid request body or missing credentials"
// @Failure 401 {object} APIResponse "Invalid credentials"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/login [post]
func AdminLogin(c *fiber.Ctx) error {
	var req AdminLoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Username and password are required",
		})
	}

	// Find admin by username
	var admin models.Admin
	if err := db.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	// Verify password
	if !admin.CheckPassword(req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	// Increment token version to invalidate all previous tokens
	// This ensures only the latest login session is valid
	admin.TokenVersion++
	if err := db.DB.Save(&admin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to update admin token version",
		})
	}

	// Generate permanent admin token (no expiry) with new token version
	token, err := utils.GenerateAdminToken(admin.ID, admin.Username, admin.Role, admin.TokenVersion)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Login successful",
		Data: fiber.Map{
			"id":     admin.ID,
			"username":     admin.Username,
			"role":         admin.Role,
			"access_token": token,
		},
	})
}
