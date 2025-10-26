package handlers

import (
	"log"
	"ololo-gate/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetLocations godoc
// @Summary Get all locations accessible to the current user
// @Description Fetch all locations from third-party API based on user's phone with their gates
// @Tags Gate Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} LocationsListResponse "Locations retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/locations [get]
func GetLocations(c *fiber.Ctx) error {
	// Get user phone from context (set by JWT middleware)
	phone, ok := c.Locals("phone").(string)
	if !ok {
		phone = "unknown"
	}

	log.Printf("Fetching locations for phone: %s", phone)

	client := services.NewThirdPartyClient()
	locations, err := client.GetAllLocationsWithGates(phone)
	if err != nil {
		log.Printf("Error fetching locations from third-party API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to fetch locations",
		})
	}

	// Convert to DTOs (include gates)
	var dtos []LocationDTO
	for _, loc := range locations {
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

		dtos = append(dtos, LocationDTO{
			ID:      loc.ID,
			Title:   loc.Title,
			Address: loc.Address,
			Logo:    loc.Logo,
			Gates:   gateDTOs,
		})
	}

	return c.Status(fiber.StatusOK).JSON(LocationsListResponse{
		Success: true,
		Message: "Locations retrieved successfully",
		Data:    dtos,
	})
}

// GetGatesByLocation godoc
// @Summary Get all gates for a specific location
// @Description Fetch all gates accessible to the current user for a specific location from third-party API
// @Tags Gate Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param locationId path int true "Location ID"
// @Success 200 {object} GatesListResponse "Gates retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid location ID"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/locations/{locationId}/gates [get]
func GetGatesByLocation(c *fiber.Ctx) error {
	locationIDStr := c.Params("locationId")
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil || locationID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid location ID",
		})
	}

	// Get user phone from context (set by JWT middleware)
	phone, ok := c.Locals("phone").(string)
	if !ok {
		phone = "unknown"
	}

	log.Printf("Fetching gates for location %d for phone: %s", locationID, phone)

	client := services.NewThirdPartyClient()
	gates, err := client.GetGatesByPhoneAndLocation(phone, locationID)
	if err != nil {
		log.Printf("Error fetching gates from third-party API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to fetch gates",
		})
	}

	// Convert to DTOs
	var dtos []GateDTO
	for _, gate := range gates {
		dtos = append(dtos, GateDTO{
			ID:               gate.ID,
			LocationID:       gate.LocationID,
			Title:            gate.Title,
			Description:      gate.Description,
			GateIsHorizontal: gate.GateIsHorizontal,
			IsOpen:           gate.IsOpen,
		})
	}

	return c.Status(fiber.StatusOK).JSON(GatesListResponse{
		Success: true,
		Message: "Gates retrieved successfully",
		Data:    dtos,
	})
}

// OpenGate godoc
// @Summary Open a gate
// @Description Send command to open a specific gate to third-party API
// @Tags Gate Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param gateId path int true "Gate ID"
// @Success 200 {object} GateActionResponse "Gate operation response"
// @Failure 400 {object} APIResponse "Invalid gate ID"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/locations/{gateId}/open [put]
func OpenGate(c *fiber.Ctx) error {
	gateIDStr := c.Params("gateId")
	gateID, err := strconv.Atoi(gateIDStr)
	if err != nil || gateID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid gate ID",
		})
	}

	// Get user phone from context
	phone, ok := c.Locals("phone").(string)
	if !ok {
		phone = "unknown"
	}

	log.Printf("User %s attempting to open gate %d", phone, gateID)

	client := services.NewThirdPartyClient()
	success, err := client.OpenGate(gateID)
	if err != nil {
		log.Printf("Error opening gate from third-party API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to open gate",
		})
	}

	response := GateActionResponse{
		Success: true,
		Message: "Gate operation completed",
		Data: GateActionData{
			GateID: gateID,
			Status: success,
		},
	}

	log.Printf("OpenGate response for gate %d: Success=%v, Status=%v", gateID, response.Success, response.Data.Status)

	return c.Status(fiber.StatusOK).JSON(response)
}

// CloseGate godoc
// @Summary Close a gate
// @Description Send command to close a specific gate to third-party API
// @Tags Gate Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param gateId path int true "Gate ID"
// @Success 200 {object} GateActionResponse "Gate operation response"
// @Failure 400 {object} APIResponse "Invalid gate ID"
// @Failure 401 {object} APIResponse "Unauthorized - invalid or missing token"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /api/v1/locations/{gateId}/close [put]
func CloseGate(c *fiber.Ctx) error {
	gateIDStr := c.Params("gateId")
	gateID, err := strconv.Atoi(gateIDStr)
	if err != nil || gateID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Message: "Invalid gate ID",
		})
	}

	// Get user phone from context
	phone, ok := c.Locals("phone").(string)
	if !ok {
		phone = "unknown"
	}

	log.Printf("User %s attempting to close gate %d", phone, gateID)

	client := services.NewThirdPartyClient()
	success, err := client.CloseGate(gateID)
	if err != nil {
		log.Printf("Error closing gate from third-party API: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Message: "Failed to close gate",
		})
	}

	response := GateActionResponse{
		Success: true,
		Message: "Gate operation completed",
		Data: GateActionData{
			GateID: gateID,
			Status: success,
		},
	}

	log.Printf("CloseGate response for gate %d: Success=%v, Status=%v", gateID, response.Success, response.Data.Status)

	return c.Status(fiber.StatusOK).JSON(response)
}
