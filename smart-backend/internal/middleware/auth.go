package middleware

import (
	"log"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTProtected is a middleware that validates JWT access tokens
func JWTProtected() fiber.Handler {
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

		// Validate the token
		claims, err := utils.ValidateToken(tokenString, utils.AccessToken)
		if err != nil {
			log.Printf("[TOKEN_VALIDATION] Invalid or expired access token: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired token",
			})
		}

		log.Printf("[TOKEN_VALIDATION] Access token validated. User ID from claims: %s, Phone: %s, Claims token_version: %d",
			claims.UserID, claims.Phone, claims.TokenVersion)

		// Verify token version against database
		var user models.User
		if err := db.DB.Select("id", "token_version").First(&user, claims.UserID).Error; err != nil {
			log.Printf("[TOKEN_VALIDATION] User ID %s not found in database: %v", claims.UserID, err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}

		log.Printf("[TOKEN_VALIDATION] User found in DB. User ID: %s, DB token_version: %d, Claims token_version: %d",
			user.ID, user.TokenVersion, claims.TokenVersion)

		// Check if token version matches
		if user.TokenVersion != claims.TokenVersion {
			log.Printf("[TOKEN_INVALIDATED] Token version mismatch for user ID %s (phone: %s). Token invalidated. Claims version=%d, DB version=%d",
				user.ID, claims.Phone, claims.TokenVersion, user.TokenVersion)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token has been invalidated. Please login again.",
			})
		}

		log.Printf("[TOKEN_VALID] Access token valid for user ID=%s (phone=%s) with token_version=%d",
			user.ID, claims.Phone, user.TokenVersion)

		// Store user info in context for use in handlers
		c.Locals("id", claims.UserID)
		c.Locals("phone", claims.Phone)

		return c.Next()
	}
}
