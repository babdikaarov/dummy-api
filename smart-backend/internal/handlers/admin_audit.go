package handlers

import (
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"

	"github.com/gofiber/fiber/v2"
)

// GetAdminAuditLogs godoc
// @Summary Get admin audit logs
// @Description Retrieve audit logs of admin actions (super admin only). Returns paginated list of all administrative operations.
// @Tags Admin Audit Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param admin_id query string false "Filter by admin ID"
// @Param action query string false "Filter by action type"
// @Param resource_type query string false "Filter by resource type"
// @Success 200 {object} PaginatedAuditLogResponse "Audit logs retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 403 {object} APIResponse "Forbidden - super admin access required"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/audit-logs [get]
func GetAdminAuditLogs(c *fiber.Ctx) error {
	// Parse pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query with filters
	query := db.DB

	// Filter by admin ID if provided
	if adminID := c.Query("admin_id"); adminID != "" {
		query = query.Where("admin_id = ?", adminID)
	}

	// Filter by action if provided
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	// Filter by resource type if provided
	if resourceType := c.Query("resource_type"); resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	// Get total count
	var total int64
	query.Model(&models.AdminAuditLog{}).Count(&total)

	// Fetch paginated results (order by most recent first)
	var logs []models.AdminAuditLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to retrieve audit logs",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Audit logs retrieved successfully",
		"data":    logs,
		"pagination": fiber.Map{
			"total":        total,
			"page":         page,
			"limit":        limit,
			"pages":        (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetAdminAuditLogByID godoc
// @Summary Get audit log by ID
// @Description Retrieve a specific audit log entry by ID (super admin only)
// @Tags Admin Audit Logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Audit log ID (UUID)"
// @Success 200 {object} AuditLogDetailResponse "Audit log retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 403 {object} APIResponse "Forbidden - super admin access required"
// @Failure 404 {object} APIResponse "Audit log not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/admin/audit-logs/{id} [get]
func GetAdminAuditLogByID(c *fiber.Ctx) error {
	logID := c.Params("id")

	var log models.AdminAuditLog
	if err := db.DB.First(&log, "id = ?", logID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Message: "Audit log not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Audit log retrieved successfully",
		"data":    log,
	})
}

// PaginatedAuditLogResponse defines the response structure for audit log list
// @name PaginatedAuditLogResponse
type PaginatedAuditLogResponse struct {
	Success    bool                    `json:"success" example:"true"`
	Message    string                  `json:"message" example:"Audit logs retrieved successfully"`
	Data       []models.AdminAuditLog  `json:"data"`
	Pagination PaginationMeta          `json:"pagination"`
}

// AuditLogDetailResponse defines the response structure for a single audit log
// @name AuditLogDetailResponse
type AuditLogDetailResponse struct {
	Success bool                  `json:"success" example:"true"`
	Message string                `json:"message" example:"Audit log retrieved successfully"`
	Data    models.AdminAuditLog  `json:"data"`
}
