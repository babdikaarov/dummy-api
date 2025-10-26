package handlers

import (
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// CreateAdminRequest defines the structure for creating a new admin
// @name CreateAdminRequest
type CreateAdminRequest struct {
	Username string `json:"username" validate:"required" example:"newadmin"`
	Password string `json:"password" validate:"required,min=6" example:"password123"`
	Role     string `json:"role" validate:"required" example:"regular"` // "super" or "regular"
}

// UpdateAdminRequest defines the structure for updating admin details (password, username, role)
// @name UpdateAdminRequest
type UpdateAdminRequest struct {
	Password *string `json:"password,omitempty" validate:"omitempty,min=6" example:"newpassword123"`
	Username *string `json:"username,omitempty" validate:"omitempty" example:"newusername"`
	Role     *string `json:"role,omitempty" validate:"omitempty" example:"regular"`
}

// GetAllAdmins godoc
// @Summary Get all admin users
// @Description Retrieve a list of all admin accounts with pagination, search, filtering, and ordering (super admin only)
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Records per page (default: 500)"
// @Param search query string false "Search by username"
// @Param role query string false "Filter by role (super or regular)"
// @Param order query string false "Order results by created_at (ASC or DESC, default: DESC)"
// @Success 200 {object} AdminsListResponse "Admin users retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} APIResponse "Forbidden - super admin access required"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/users [get]
func GetAllAdmins(c *fiber.Ctx) error {
	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 500)
	search := c.Query("search", "")
	roleFilter := c.Query("role", "")
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
	query := db.DB.Select("id", "username", "role", "created_at", "updated_at")

	// Apply search filter
	if search != "" {
		query = query.Where("username LIKE ?", "%"+search+"%")
	}

	// Apply role filter
	if roleFilter != "" {
		if roleFilter != models.RoleSuper && roleFilter != models.RoleRegular {
			return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
				Success: false,
				Message: "Invalid role. Must be 'super' or 'regular'",
			})
		}
		query = query.Where("role = ?", roleFilter)
	}

	// Apply order
	query = query.Order("created_at " + order)

	// Get total count before pagination
	var total int64
	if err := query.Model(&models.Admin{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to retrieve admins",
		})
	}

	// Apply pagination
	if limit != -1 {
		offset := (page - 1) * limit
		query = query.Offset(offset).Limit(limit)
	}

	// Fetch admins
	var admins []models.Admin
	if err := query.Find(&admins).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to retrieve admins",
		})
	}

	// Map admins to AdminDTO
	adminDTOs := make([]AdminDTO, len(admins))
	for i, admin := range admins {
		adminDTOs[i] = AdminDTO{
			ID:        admin.ID,
			Username:  admin.Username,
			Role:      admin.Role,
			CreatedAt: admin.CreatedAt,
			UpdatedAt: admin.UpdatedAt,
		}
	}

	// Calculate pagination metadata
	perPage := len(admins)
	if limit != -1 {
		perPage = limit
	} else {
		perPage = int(total)
	}

	lastPage := 1
	if limit != -1 && perPage > 0 {
		lastPage = int((total + int64(limit) - 1) / int64(limit))
	}

	return c.Status(fiber.StatusOK).JSON(AdminsListResponse{
		Success: true,
		Message: "Admins retrieved successfully",
		Data:    adminDTOs,
		Pagination: PaginationMeta{
			Total:       int(total),
			PerPage:     perPage,
			CurrentPage: page,
			LastPage:    lastPage,
		},
	})
}

// CreateAdmin godoc
// @Summary Create a new admin user
// @Description Create a new admin account with specified role (super admin only)
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateAdminRequest true "Admin creation details"
// @Success 201 {object} AdminResponse "Admin user created successfully"
// @Failure 400 {object} APIResponse "Invalid request body or validation error"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} APIResponse "Forbidden - super admin access required"
// @Failure 409 {object} APIResponse "Admin with this username already exists"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/users [post]
func CreateAdmin(c *fiber.Ctx) error {
	var req CreateAdminRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate role
	if req.Role != models.RoleSuper && req.Role != models.RoleRegular {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid role. Must be 'super' or 'regular'",
		})
	}

	// Validate password length
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		})
	}

	// Check if admin with this username already exists
	var existingAdmin models.Admin
	if err := db.DB.Where("username = ?", req.Username).First(&existingAdmin).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(APIResponse{
			Success: false,
			Message: "Admin with this username already exists",
		})
	}

	// Create new admin (password will be hashed by BeforeCreate hook)
	admin := models.Admin{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
	}

	if err := db.DB.Create(&admin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to create admin",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Message: "Admin created successfully",
		Data: fiber.Map{
			"id": admin.ID,
			"username": admin.Username,
			"role":     admin.Role,
		},
	})
}

// GetAdminByID godoc
// @Summary Get admin by ID
// @Description Retrieve a specific admin's details by ID. Super admins can retrieve any admin. Regular admins can only retrieve their own details.
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID (UUID)"
// @Success 200 {object} AdminDetailResponse "Admin retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid admin ID format"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} APIResponse "Forbidden - regular admins can only access their own record"
// @Failure 404 {object} APIResponse "Admin not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/users/{id} [get]
func GetAdminByID(c *fiber.Ctx) error {
	// Get admin ID from URL parameter
	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid admin ID format",
		})
	}

	// Get current admin role and ID from context
	requestingAdminRole := c.Locals("admin_role").(string)
	requestingAdminID := c.Locals("id").(uuid.UUID)

	// Check access: regular admins can only access their own record
	if requestingAdminRole != models.RoleSuper && requestingAdminID != adminID {
		return c.Status(fiber.StatusForbidden).JSON(APIResponse{
			Success: false,
			Message: "Regular admins can only access their own record",
		})
	}

	// Find admin
	var admin models.Admin
	if err := db.DB.First(&admin, adminID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "Admin not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(AdminDetailResponse{
		Success: true,
		Message: "Admin retrieved successfully",
		Data: AdminDetailData{
			AdminID:   admin.ID,
			Username:  admin.Username,
			Role:      admin.Role,
			CreatedAt: admin.CreatedAt,
			UpdatedAt: admin.UpdatedAt,
		},
	})
}

// UpdateAdmin godoc
// @Summary Update admin details
// @Description Update an admin's details (password, username, and/or role). Super admins can update any admin. Regular admins can only update their own password and username (not role).
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID (UUID)"
// @Param request body UpdateAdminRequest true "Update details (at least one field required)"
// @Success 200 {object} AdminResponse "Admin updated successfully"
// @Failure 400 {object} APIResponse "Invalid admin ID or request body"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} APIResponse "Forbidden - insufficient permissions for this operation"
// @Failure 404 {object} APIResponse "Admin not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/users/{id} [patch]
func UpdateAdmin(c *fiber.Ctx) error {
	// Get admin ID from URL parameter
	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid admin ID format",
		})
	}

	// Get current admin role and ID from context
	requestingAdminRole := c.Locals("admin_role").(string)
	requestingAdminID := c.Locals("id").(uuid.UUID)

	// Check if trying to update different admin as regular admin
	isUpdatingDifferentAdmin := requestingAdminID != adminID
	if isUpdatingDifferentAdmin && requestingAdminRole != models.RoleSuper {
		return c.Status(fiber.StatusForbidden).JSON(APIResponse{
			Success: false,
			Message: "Regular admins can only update their own record",
		})
	}

	var req UpdateAdminRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate at least one field is provided
	if req.Password == nil && req.Username == nil && req.Role == nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "At least one field (password, username, or role) must be provided",
		})
	}

	// Regular admin trying to update role
	if req.Role != nil && requestingAdminRole != models.RoleSuper {
		return c.Status(fiber.StatusForbidden).JSON(APIResponse{
			Success: false,
			Message: "Only super admins can change admin roles",
		})
	}

	// Find admin
	var admin models.Admin
	if err := db.DB.First(&admin, adminID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "Admin not found",
		})
	}

	// Update password if provided
	if req.Password != nil {
		if len(*req.Password) < 6 {
			return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
				Success: false,
				Message: "Password must be at least 6 characters long",
			})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
				Success: false,
				Message: "Failed to hash password",
			})
		}
		admin.Password = string(hashedPassword)
	}

	// Update username if provided
	if req.Username != nil {
		admin.Username = *req.Username
	}

	// Update role if provided (only super admin can do this)
	if req.Role != nil {
		if *req.Role != models.RoleSuper && *req.Role != models.RoleRegular {
			return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
				Success: false,
				Message: "Invalid role. Must be 'super' or 'regular'",
			})
		}
		admin.Role = *req.Role
	}

	// Save changes
	if err := db.DB.Save(&admin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to update admin",
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Admin updated successfully",
		Data: fiber.Map{
			"id":       admin.ID,
			"username": admin.Username,
			"role":     admin.Role,
		},
	})
}

// DeleteAdmin godoc
// @Summary Delete an admin user
// @Description Delete an admin account by ID (soft delete, super admin only)
// @Tags Admin User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Admin ID (UUID)"
// @Success 200 {object} AdminResponse "Admin user deleted successfully"
// @Failure 400 {object} APIResponse "Invalid admin ID format"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} APIResponse "Forbidden - super admin access required"
// @Failure 404 {object} APIResponse "Admin not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/users/{id} [delete]
func DeleteAdmin(c *fiber.Ctx) error {
	// Get admin ID from URL parameter
	adminID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid admin ID format",
		})
	}

	// Prevent deletion of initial super admin
	initialAdminUUID, err := uuid.Parse(db.DB.Config.Name())
	if err == nil && adminID == initialAdminUUID {
		return c.Status(fiber.StatusForbidden).JSON(APIResponse{
			Success: false,
			Message: "Cannot delete the initial super admin",
		})
	}

	// Find admin
	var admin models.Admin
	if err := db.DB.First(&admin, adminID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "Admin not found",
		})
	}

	// Delete admin (soft delete)
	if err := db.DB.Delete(&admin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to delete admin",
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Success: true,
		Message: "Admin deleted successfully",
		Data: fiber.Map{
			"id": admin.ID,
			"username": admin.Username,
		},
	})
}
