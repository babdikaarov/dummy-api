package middleware

import (
	"log"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AdminJWTProtected validates admin JWT tokens and checks token version
func AdminJWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Missing authorization header",
			})
		}

		// Check if it starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header format. Use: Bearer <token>",
			})
		}

		tokenString := parts[1]

		// Validate the admin token
		claims, err := utils.ValidateAdminToken(tokenString)
		if err != nil {
			log.Printf("[ADMIN_TOKEN_VALIDATION] Invalid or expired admin token: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
			})
		}

		log.Printf("[ADMIN_TOKEN_VALIDATION] Admin token validated. Admin ID from claims: %s, Username: %s, Claims token_version: %d",
			claims.AdminID, claims.Username, claims.TokenVersion)

		// Check if token version matches the database
		// This invalidates tokens when admin logs in from another device
		var admin models.Admin
		if err := db.DB.First(&admin, claims.AdminID).Error; err != nil {
			log.Printf("[ADMIN_TOKEN_VALIDATION] Admin ID %s not found in database: %v", claims.AdminID, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token has been invalidated",
			})
		}

		log.Printf("[ADMIN_TOKEN_VALIDATION] Admin found in DB. Admin ID: %s, DB token_version: %d, Claims token_version: %d",
			admin.ID, admin.TokenVersion, claims.TokenVersion)

		if admin.TokenVersion != claims.TokenVersion {
			log.Printf("[ADMIN_TOKEN_INVALIDATED] Token version mismatch for admin ID %s (username: %s). Token invalidated. Claims version=%d, DB version=%d",
				admin.ID, claims.Username, claims.TokenVersion, admin.TokenVersion)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token has been invalidated",
			})
		}

		log.Printf("[ADMIN_TOKEN_VALID] Admin token valid for admin ID=%s (username=%s) with token_version=%d",
			admin.ID, claims.Username, admin.TokenVersion)

		// Store admin info in context for use in handlers
		c.Locals("id", claims.AdminID)
		c.Locals("admin_username", claims.Username)
		c.Locals("admin_role", claims.Role)

		return c.Next()
	}
}

// SuperAdminOnly middleware checks if the admin has super admin role
func SuperAdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get admin role from context (must run AdminJWTProtected first)
		role := c.Locals("admin_role")

		if role == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authentication required",
			})
		}

		if role != models.RoleSuper {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Super admin access required",
			})
		}

		return c.Next()
	}
}
