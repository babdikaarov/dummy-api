package handlers

import (
	"log"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"

	"github.com/gofiber/fiber/v2"
)

// UpdateContactRequest defines the structure for updating contact information
// @name UpdateContactRequest
type UpdateContactRequest struct {
	SupportNumber int    `json:"support_number" validate:"required" example:"77091234567"`
	EmailSupport  string `json:"email_support" validate:"required,email" example:"support@ololo.com"`
	Address       string `json:"address" validate:"required" example:"г. Бишкек, проспект Чуй, 135"`
}

// GetContact godoc
// @Summary Get contact information
// @Description Retrieve the application's contact information (public endpoint, no authentication required). Returns empty values if contact information has not been set.
// @Tags Contact Information
// @Accept json
// @Produce json
// @Success 200 {object} ContactResponse "Contact information retrieved successfully"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/contacts [get]
func GetContact(c *fiber.Ctx) error {
	var contact models.Contact

	// Try to fetch the first (and should be only) contact record
	// If not found, return empty values with status 200
	if err := db.DB.First(&contact).Error; err != nil {
		log.Println("No contact information found, returning empty values")
		return c.Status(fiber.StatusOK).JSON(ContactResponse{
			Success: true,
			Message: "Contact information retrieved successfully",
			Data: ContactDTO{
				SupportNumber: 0,
				EmailSupport:  "",
				Address:       "",
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(ContactResponse{
		Success: true,
		Message: "Contact information retrieved successfully",
		Data: ContactDTO{
			SupportNumber: contact.SupportNumber,
			EmailSupport:  contact.EmailSupport,
			Address:       contact.Address,
		},
	})
}

// UpdateContact godoc
// @Summary Update contact information
// @Description Update or create the application's contact information (admin only). Creates a new contact record if one doesn't exist.
// @Tags Contact Information
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateContactRequest true "Contact information to update"
// @Success 200 {object} ContactResponse "Contact information updated or created successfully"
// @Failure 400 {object} APIResponse "Invalid request body or validation error"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing admin token"
// @Failure 403 {object} APIResponse "Forbidden - admin access required"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/contacts [patch]
func UpdateContact(c *fiber.Ctx) error {
	var req UpdateContactRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Validate support number (basic validation - should be a valid phone number)
	if req.SupportNumber <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Support number must be a valid phone number",
		})
	}

	// Validate email support
	if req.EmailSupport == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Email support is required",
		})
	}

	// Validate address
	if req.Address == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Address is required",
		})
	}

	// Try to fetch the first contact record
	var contact models.Contact
	if err := db.DB.First(&contact).Error; err != nil {
		// If not found, create a new contact record
		contact = models.Contact{
			SupportNumber: req.SupportNumber,
			EmailSupport:  req.EmailSupport,
			Address:       req.Address,
		}
		if err := db.DB.Create(&contact).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
				Success: false,
				Message: "Failed to create contact information",
			})
		}
	} else {
		// Update existing contact record
		contact.SupportNumber = req.SupportNumber
		contact.EmailSupport = req.EmailSupport
		contact.Address = req.Address

		if err := db.DB.Save(&contact).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
				Success: false,
				Message: "Failed to update contact information",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(ContactResponse{
		Success: true,
		Message: "Contact information updated successfully",
		Data: ContactDTO{
			SupportNumber: contact.SupportNumber,
			EmailSupport:  contact.EmailSupport,
			Address:       contact.Address,
		},
	})
}
