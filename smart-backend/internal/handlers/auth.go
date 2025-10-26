package handlers

import (
	"log"
	"ololo-gate/internal/config"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/utils"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

// RegisterRequest defines the structure for registration requests
// @name RegisterRequest
type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required" example:"+77771234567"`
	Password string `json:"password" validate:"required,min=6" example:"password123"`
}

// LoginRequest defines the structure for login requests
// @name LoginRequest
type LoginRequest struct {
	Phone    string `json:"phone" validate:"required" example:"+77771234567"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// RefreshRequest defines the structure for token refresh requests
// @name RefreshRequest
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// APIResponse is a standard response format
// @name APIResponse
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Phone number validation regex (E.164 format)
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// Register godoc
// @Summary Register a new user
// @Description Register a new user account with phone number and password (E.164 format required)
// @Tags User Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} RegisterResponse "User registered successfully"
// @Failure 400 {object} APIResponse "Invalid request body or validation error"
// @Failure 409 {object} APIResponse "User with this phone number already exists"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/auth/register [post]
func Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate phone number format
	if !phoneRegex.MatchString(req.Phone) {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid phone number format. Use international format (e.g., +77771234567)",
		})
	}

	// Validate password length
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		})
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.DB.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(APIResponse{
			Success: false,
			Message: "User with this phone number already exists",
		})
	}

	// Create new user (password will be hashed by BeforeCreate hook)
	user := models.User{
		Phone:    req.Phone,
		Password: req.Password,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data: fiber.Map{
			"id": user.ID,
			"phone":   user.Phone,
		},
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user with phone and password, returns access and refresh tokens. Supports device-based token invalidation.
// @Tags User Authentication
// @Accept json
// @Produce json
// @Param device_id query string false "Unique device identifier (optional - if provided and different from current device, previous tokens will be invalidated)"
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful with tokens"
// @Failure 400 {object} APIResponse "Invalid request body or phone format"
// @Failure 401 {object} APIResponse "Invalid credentials"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func Login(c *fiber.Ctx) error {
	var req LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate phone number format
	if !phoneRegex.MatchString(req.Phone) {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid phone number format",
		})
	}

	// Find user by phone
	var user models.User
	log.Printf("[LOGIN] Attempting login with phone: %s", req.Phone)
	if err := db.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		log.Printf("[LOGIN_FAILED] Phone %s not found in database: %v", req.Phone, err)
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	log.Printf("[LOGIN] User found in database: ID=%s, Phone=%s, DB token_version=%d", user.ID, user.Phone, user.TokenVersion)

	// Verify password
	if !user.CheckPassword(req.Password) {
		log.Printf("[LOGIN_FAILED] Password verification FAILED for user ID=%s (phone=%s). Provided password hash did not match stored hash.", user.ID, user.Phone)
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	log.Printf("[LOGIN] Password verification SUCCESSFUL for user ID=%s (phone=%s)", user.ID, user.Phone)

	// Get optional device_id from query parameters (accept both deviceId and device_id)
	deviceID := c.Query("deviceId")
	if deviceID == "" {
		deviceID = c.Query("device_id")
	}

	log.Printf("[LOGIN] Device tracking: provided=%s, current=%s", deviceID, user.CurrentDeviceID)

	// Determine if device changed and whether to increment token version
	// Device change logic:
	// - If device_id not provided: increment token_version (backward compatibility, old behavior)
	// - If device_id provided and different from current: increment token_version (new device)
	// - If device_id provided and same as current: don't increment (same device, reuse session)
	deviceChanged := false
	previousDeviceID := user.CurrentDeviceID

	if deviceID == "" {
		// No device_id provided: increment token_version for backward compatibility
		deviceChanged = true
		log.Printf("[LOGIN] No device_id provided. Will increment token_version for backward compatibility.")
	} else {
		// Device_id provided: check if it's different from current
		deviceChanged = user.CurrentDeviceID != "" && user.CurrentDeviceID != deviceID
		if deviceChanged {
			log.Printf("[LOGIN] Device CHANGED: old=%s, new=%s. Will increment token_version.", user.CurrentDeviceID, deviceID)
		} else {
			log.Printf("[LOGIN] Device SAME: %s. Will NOT increment token_version.", deviceID)
		}
	}

	// Increment token version only if device changed
	oldTokenVersion := user.TokenVersion
	if deviceChanged {
		user.TokenVersion++
		log.Printf("[LOGIN] Token version incremented: %d -> %d", oldTokenVersion, user.TokenVersion)
	}

	// Update current device ID if device_id provided
	if deviceID != "" {
		user.CurrentDeviceID = deviceID
	}

	if err := db.DB.Save(&user).Error; err != nil {
		log.Printf("[LOGIN_FAILED] Failed to save user token_version update: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to update user token version",
		})
	}

	// Log device change event for audit purposes (backend only, not sent to client)
	if deviceChanged && deviceID != "" {
		log.Printf("[DEVICE_CHANGE] User: %s (ID: %s) changed device from '%s' to '%s'",
			user.Phone, user.ID, previousDeviceID, deviceID)
	}

	// Generate tokens with current token version
	tokens, err := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)
	if err != nil {
		log.Printf("[LOGIN_FAILED] Failed to generate tokens: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to generate tokens",
		})
	}

	log.Printf("[LOGIN_SUCCESS] Login successful for user ID=%s (phone=%s). Tokens generated with token_version=%d, device_id=%s",
		user.ID, user.Phone, user.TokenVersion, deviceID)

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Login successful",
		Data: fiber.Map{
			"id":                  user.ID,
			"phone":               user.Phone,
			"access_token":        tokens.AccessToken,
			"refresh_token":       tokens.RefreshToken,
			"access_expires_in":   int64(config.AppConfig.JWT.AccessExpiry.Seconds()),
			"refresh_expires_in":  int64(config.AppConfig.JWT.RefreshExpiry.Seconds()),
		},
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags User Authentication
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} RefreshResponse "New access token generated"
// @Failure 400 {object} APIResponse "Invalid request body"
// @Failure 401 {object} APIResponse "Invalid or expired refresh token, or token has been invalidated"
// @Failure 404 {object} APIResponse "User not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func RefreshToken(c *fiber.Ctx) error {
	var req RefreshRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate refresh token and get claims
	claims, err := utils.ValidateToken(req.RefreshToken, utils.RefreshToken)

	if err != nil {
		log.Printf("[REFRESH_FAILED] Invalid or expired refresh token: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "Invalid or expired refresh token",
		})
	}

	log.Printf("[REFRESH] Refresh token received. User ID from claims: %s, Claims token_version: %d", claims.UserID, claims.TokenVersion)

	// Verify token version against database
	var user models.User
	if err := db.DB.Select("id", "token_version").First(&user, claims.UserID).Error; err != nil {
		log.Printf("[REFRESH_FAILED] User ID %s not found in database: %v", claims.UserID, err)
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "User not found",
		})
	}

	log.Printf("[REFRESH] User found in database: User ID=%s, DB token_version=%d, Claims token_version=%d", user.ID, user.TokenVersion, claims.TokenVersion)

	// Check if token version matches
	if user.TokenVersion != claims.TokenVersion {
		log.Printf("[REFRESH_FAILED] Token version mismatch for user ID %s. Token invalidated. Claims version=%d, DB version=%d",
			user.ID, claims.TokenVersion, user.TokenVersion)
		return c.Status(fiber.StatusUnauthorized).JSON(APIResponse{
			Success: false,
			Message: "Token has been invalidated. Please login again.",
		})
	}

	log.Printf("[REFRESH] Token version match verified. Generating new access token for user ID=%s", user.ID)

	// Generate new access token from refresh token
	accessToken, err := utils.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		log.Printf("[REFRESH_FAILED] Failed to generate new access token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to generate access token",
		})
	}

	log.Printf("[REFRESH_SUCCESS] New access token generated for user ID=%s with token_version=%d", user.ID, user.TokenVersion)

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data: fiber.Map{
			"access_token": accessToken,
		},
	})
}

// CheckPhoneAvailability godoc
// @Summary Check if phone number is available for registration
// @Description Check if a phone number is available for registration or account creation (public endpoint, no authentication required)
// @Tags User Authentication
// @Accept json
// @Produce json
// @Param phone query string true "Phone number in E.164 format (e.g., +77771234567)"
// @Success 200 {object} PhoneAvailabilityResponse "Phone availability check result"
// @Failure 400 {object} APIResponse "Invalid phone number format"
// @Router /api/v1/auth/check-phone [get]
func CheckPhoneAvailability(c *fiber.Ctx) error {
	phone := c.Query("phone")

	// Validate phone number is provided
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Phone number is required",
		})
	}

	// Validate phone number format
	if !phoneRegex.MatchString(phone) {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid phone number format. Use international format (e.g., +77771234567)",
		})
	}

	// Check if phone number exists
	var existingUser models.User
	isAvailable := true
	if err := db.DB.Where("phone = ?", phone).First(&existingUser).Error; err == nil {
		// Phone number exists - not available
		isAvailable = false
	}

	return c.Status(fiber.StatusOK).JSON(PhoneAvailabilityResponse{
		Success:   true,
		Message:   "Phone availability checked",
		Available: isAvailable,
	})
}
