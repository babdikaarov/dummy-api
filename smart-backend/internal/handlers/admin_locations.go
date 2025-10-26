package handlers

import (
	"log"
	"ololo-gate/internal/services"

	"github.com/gofiber/fiber/v2"
)

// GetAvailableLocations godoc
// @Summary Get all available locations in the system
// @Description Fetch all locations from third-party API without filtering by user (admin access only)
// @Tags Location Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AvailableLocationsResponse "Available locations retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 403 {object} APIResponse "Forbidden - requires admin access"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/available-locations [get]
func GetAvailableLocations(c *fiber.Ctx) error {
	// JWT middleware ensures admin is authenticated
	adminUsername, ok := c.Locals("admin_username").(string)
	if !ok {
		adminUsername = "unknown"
	}

	log.Printf("Admin %s fetching all available locations", adminUsername)

	client := services.NewThirdPartyClient()
	locations, err := client.GetAllLocations()
	if err != nil {
		log.Printf("Error fetching locations from third-party API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to fetch locations from third-party API",
		})
	}

	log.Printf("Fetched %d locations from third-party API", len(locations))

	// Convert to DTOs (include gates)
	var dtos []LocationDTO
	for _, loc := range locations {

		// Initialize gates as empty array to avoid null serialization
		gateDTOs := make([]GateDTO, 0)
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

		dtos = append(dtos, LocationDTO{
			ID:      loc.ID,
			Title:   loc.Title,
			Address: loc.Address,
			Logo:    loc.Logo,
			Gates:   gateDTOs,
		})
	}

	return c.Status(fiber.StatusOK).JSON(AvailableLocationsResponse{
		Success: true,
		Message: "Available locations retrieved successfully",
		Data:    dtos,
	})
}
