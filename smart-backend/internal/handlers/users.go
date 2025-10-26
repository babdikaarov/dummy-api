package handlers

import (
	"encoding/json"
	"log"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/services"
	"ololo-gate/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// GetAllUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all registered users with pagination and search (requires admin authentication)
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Records per page (default: 500)"
// @Param search query string false "Search by phone number"
// @Param order query string false "Order results by created_at (ASC or DESC, default: DESC)"
// @Success 200 {object} UsersListResponse "Users retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/users [get]
func GetAllUsers(c *fiber.Ctx) error {
	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 500)
	search := c.Query("search", "")
	order := c.Query("order", "DESC")

	// Validate page
	if page < 1 {
		page = 1
	}

	// Validate limit
	if limit != -1 && limit < 1 {
		limit = 10
	}
	if limit > 500 {
		limit = 500
	}

	// Validate order parameter
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}

	// Build query
	query := db.DB.Select("id", "phone", "created_at", "updated_at")

	// Apply search filter
	if search != "" {
		query = query.Where("phone LIKE ?", "%"+search+"%")
	}

	// Apply order
	query = query.Order("created_at " + order)

	// Get total count before pagination
	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to retrieve users",
		})
	}

	// Apply pagination
	if limit != -1 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Fetch users
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to retrieve users",
		})
	}

	// Map users to UserDTO
	userDTOs := make([]UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = UserDTO{
			ID:        user.ID,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	// Calculate pagination metadata
	perPage := len(users)
	if limit != -1 {
		perPage = limit
	} else {
		perPage = int(total)
	}

	lastPage := 1
	if limit != -1 && perPage > 0 {
		lastPage = int((total + int64(limit) - 1) / int64(limit))
	}

	return c.Status(fiber.StatusOK).JSON(UsersListResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    userDTOs,
		Pagination: PaginationMeta{
			Total:       int(total),
			PerPage:     perPage,
			CurrentPage: page,
			LastPage:    lastPage,
		},
	})
}

// CreateUser godoc
// @Summary Create a new user with location and gate assignment
// @Description Create a new user account and assign locations and gates via third-party API (requires admin authentication)
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "User creation details with locations and gates"
// @Success 201 {object} UserResponse "User created successfully"
// @Failure 400 {object} APIResponse "Invalid request body or validation error"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 409 {object} APIResponse "User with this phone number already exists"
// @Failure 500 {object} APIResponse "Internal server error or third-party API failure"
// @Router /api/v1/users [post]
func CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest

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

	// Location and gate IDs are optional - user can be created without them
	// and assigned later

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
		Phone:        req.Phone,
		Password:     req.Password,
		TokenVersion: 0, // Initialize token version
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to create user",
		})
	}

	log.Printf("User %s created successfully in database", req.Phone)

	// Get admin info from context
	adminUsername, ok := c.Locals("admin_username").(string)
	if !ok {
		adminUsername = "unknown"
	}
	adminID, ok := c.Locals("id").(uuid.UUID)
	if !ok {
		adminID = uuid.Nil
	}

	// Only try to assign locations and gates if they are provided
	if len(req.Locations) > 0 {
		// Transform LocationAssignmentRequest to LocationAssignmentDTO
		locations := make([]services.LocationAssignmentDTO, len(req.Locations))
		for i, loc := range req.Locations {
			locations[i] = services.LocationAssignmentDTO{
				LocationID: loc.LocationID,
				GateIds:    loc.GateIds,
			}
		}

		assignment := services.UserLocationGateAssignmentDTO{
			Phone:     req.Phone,
			Locations: locations,
		}

		client := services.NewThirdPartyClient()
		err := client.AssignUserToLocationsAndGates(assignment)

		// Log audit event
		auditDetails, _ := json.Marshal(fiber.Map{
			"phone":     req.Phone,
			"locations": req.Locations,
		})

		// Option B: Keep user in DB but return warning if assignment fails
		if err != nil {
			log.Printf("Warning: Failed to assign locations/gates to user %s (admin: %s): %v", req.Phone, adminUsername, err)
			utils.LogAdminAction(
				adminID,
				adminUsername,
				"create_user_with_assignment",
				"user",
				user.ID.String(),
				string(auditDetails),
				c.IP(),
				c.Get("User-Agent"),
				"failed",
				"Failed to assign locations/gates: "+err.Error(),
			)
			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"success": true,
				"message": "User created successfully but location assignment failed. Please try to assign locations and gates again.",
				"warning": "Third-party API assignment error: " + err.Error(),
				"data": fiber.Map{
					"id":    user.ID,
					"phone": user.Phone,
				},
			})
		}

		log.Printf("User %s created and assigned to locations/gates by admin %s", req.Phone, adminUsername)

		utils.LogAdminAction(
			adminID,
			adminUsername,
			"create_user_with_assignment",
			"user",
			user.ID.String(),
			string(auditDetails),
			c.IP(),
			c.Get("User-Agent"),
			"success",
			"",
		)
	} else {
		// User created without location/gate assignment
		utils.LogAdminAction(
			adminID,
			adminUsername,
			"create_user",
			"user",
			user.ID.String(),
			`{"phone":"`+req.Phone+`"}`,
			c.IP(),
			c.Get("User-Agent"),
			"success",
			"",
		)
	}

	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "User created successfully",
		Data: fiber.Map{
			"id":    user.ID,
			"phone": user.Phone,
		},
	})
}

// UpdateUser godoc
// @Summary Update user password and location/gate assignments
// @Description Update a user's password (optional) and reassign locations and gates via third-party API (requires admin authentication)
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body UpdateUserRequest true "Update details (password optional, locations and gates required)"
// @Success 200 {object} UserResponse "User updated successfully"
// @Failure 400 {object} APIResponse "Invalid user ID or request body"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 404 {object} APIResponse "User not found"
// @Failure 500 {object} APIResponse "Internal server error or third-party API failure"
// @Router /api/v1/users/{id} [patch]
func UpdateUser(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
	}

	var req UpdateUserRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// All fields are optional - validate only if provided
	// If password is provided, validate it
	if req.Password != "" && len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		})
	}

	// Find user
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "User not found",
		})
	}

	log.Printf("Updating user %s (phone: %s)", userID, user.Phone)

	// Get admin info from context
	adminUsername, ok := c.Locals("admin_username").(string)
	if !ok {
		adminUsername = "unknown"
	}
	adminID, ok := c.Locals("id").(uuid.UUID)
	if !ok {
		adminID = uuid.Nil
	}

	// Validate phone number if provided and different from current
	if req.Phone != "" && req.Phone != user.Phone {
		// Validate phone format
		if !phoneRegex.MatchString(req.Phone) {
			return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
				Success: false,
				Message: "Invalid phone number format. Use international format (e.g., +77771234567)",
			})
		}

		// Check if new phone number is already in use
		var existingUser models.User
		if err := db.DB.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(APIResponse{
				Success: false,
				Message: "Phone number is already in use",
			})
		}

		log.Printf("Updating phone number for user %s from %s to %s by admin %s", userID, user.Phone, req.Phone, adminUsername)
		user.Phone = req.Phone
	}

	// Build audit details
	auditDetails, _ := json.Marshal(fiber.Map{
		"phone_updated":     req.Phone != "" && req.Phone != user.Phone,
		"new_phone":         req.Phone,
		"password_updated":  req.Password != "",
		"locations":         req.Locations,
	})

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
				Success: false,
				Message: "Failed to hash password",
			})
		}

		// Update password and increment token version (this invalidates all existing tokens)
		user.Password = string(hashedPassword)
		user.TokenVersion++
		log.Printf("Password updated for user %s by admin %s", user.Phone, adminUsername)
	}

	// Increment token version if phone was changed (invalidate all existing tokens)
	if req.Phone != "" && req.Phone != user.Phone {
		user.TokenVersion++
		log.Printf("Token version incremented due to phone number change for user %s", user.Phone)
	}

	if err := db.DB.Save(&user).Error; err != nil {
		utils.LogAdminAction(
			adminID,
			adminUsername,
			"update_user",
			"user",
			user.ID.String(),
			string(auditDetails),
			c.IP(),
			c.Get("User-Agent"),
			"failed",
			"Failed to update user in database",
		)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to update user",
		})
	}

	// Only try to assign locations and gates if they are provided
	if len(req.Locations) > 0 {
		// Transform LocationAssignmentRequest to LocationAssignmentDTO
		locations := make([]services.LocationAssignmentDTO, len(req.Locations))
		for i, loc := range req.Locations {
			locations[i] = services.LocationAssignmentDTO{
				LocationID: loc.LocationID,
				GateIds:    loc.GateIds,
			}
		}

		assignment := services.UserLocationGateAssignmentDTO{
			Phone:     user.Phone,
			Locations: locations,
		}

		client := services.NewThirdPartyClient()
		err := client.AssignUserToLocationsAndGates(assignment)

		// Option B: Keep user update but return warning if assignment fails
		if err != nil {
			log.Printf("Warning: Failed to update locations/gates for user %s (admin: %s): %v", user.Phone, adminUsername, err)
			utils.LogAdminAction(
				adminID,
				adminUsername,
				"update_user_assignment",
				"user",
				user.ID.String(),
				string(auditDetails),
				c.IP(),
				c.Get("User-Agent"),
				"failed",
				"Failed to assign locations/gates: "+err.Error(),
			)
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"success": true,
				"message": "User updated successfully but location assignment failed. Please try to assign locations and gates again.",
				"warning": "Third-party API assignment error: " + err.Error(),
				"data": fiber.Map{
					"id":    user.ID,
					"phone": user.Phone,
				},
			})
		}

		log.Printf("User %s updated and assigned to locations/gates by admin %s", user.Phone, adminUsername)
		utils.LogAdminAction(
			adminID,
			adminUsername,
			"update_user_assignment",
			"user",
			user.ID.String(),
			string(auditDetails),
			c.IP(),
			c.Get("User-Agent"),
			"success",
			"",
		)
	} else {
		// User updated without assignment changes
		utils.LogAdminAction(
			adminID,
			adminUsername,
			"update_user",
			"user",
			user.ID.String(),
			string(auditDetails),
			c.IP(),
			c.Get("User-Agent"),
			"success",
			"",
		)
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "User updated successfully",
		Data: fiber.Map{
			"id":    user.ID,
			"phone": user.Phone,
		},
	})
}

// GetUserByID godoc
// @Summary Get user by ID with assigned locations and gates
// @Description Retrieve a specific user's details by ID including their assigned locations and gates from third-party API (requires admin authentication)
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} UserDetailResponse "User retrieved successfully with locations"
// @Failure 400 {object} APIResponse "Invalid user ID format"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 404 {object} APIResponse "User not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/users/{id} [get]
func GetUserByID(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
	}

	// Find user
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "User not found",
		})
	}

	log.Printf("Fetching user details for %s (ID: %s)", user.Phone, userID)

	// Fetch user's locations and gates from third-party API
	client := services.NewThirdPartyClient()
	locationsWithGates, err := client.GetAllLocationsWithGates(user.Phone)
	if err != nil {
		log.Printf("Warning: Failed to fetch locations for user %s: %v", user.Phone, err)
		// Return user info even if third-party API fails
		return c.Status(fiber.StatusOK).JSON(UserDetailResponse{
			Success: true,
			Message: "User retrieved but location data unavailable",
			Data: UserDetailDTO{
				ID:        user.ID,
				Phone:     user.Phone,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Locations: []LocationDTO{},
			},
		})
	}

	// Convert LocationResponse to LocationDTO
	var locationDTOs []LocationDTO
	for _, loc := range locationsWithGates {
		var gateDTOs []GateDTO
		for _, gate := range loc.Gates {
			gateDTOs = append(gateDTOs, GateDTO{
				ID:               gate.ID,
				Title:            gate.Title,
				Description:      gate.Description,
				LocationID:       gate.LocationID,
				IsOpen:           gate.IsOpen,
				GateIsHorizontal: gate.GateIsHorizontal,
			})
		}

		locationDTOs = append(locationDTOs, LocationDTO{
			ID:      loc.ID,
			Title:   loc.Title,
			Address: loc.Address,
			Logo:    loc.Logo,
			Gates:   gateDTOs,
		})
	}

	return c.Status(fiber.StatusOK).JSON(UserDetailResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data: UserDetailDTO{
			ID:        user.ID,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Locations: locationDTOs,
		},
	})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user account by ID (soft delete, requires admin authentication)
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} UserResponse "User deleted successfully"
// @Failure 400 {object} APIResponse "Invalid user ID format"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 404 {object} APIResponse "User not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/users/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	// Get user ID from URL parameter
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid user ID format",
		})
	}

	// Find user
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "User not found",
		})
	}

	// Invalidate all user tokens by incrementing token version
	user.TokenVersion++

	// Delete user (soft delete by default with GORM)
	if err := db.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to invalidate user tokens",
		})
	}

	if err := db.DB.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to delete user",
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "User deleted successfully",
		Data: fiber.Map{
			"id": user.ID,
			"phone":   user.Phone,
		},
	})
}
