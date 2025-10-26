package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/utils"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetAvailableLocations_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create admin
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "admin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&admin)

	token, _ := utils.GenerateAdminToken(admin.ID, admin.Username, admin.Role, 0)

	req := httptest.NewRequest("GET", "/api/v1/available-locations", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should not return unauthorized
	assert.NotEqual(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response AvailableLocationsResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Should have proper response structure
	assert.NotNil(t, response.Data)
}

func TestGetAvailableLocations_Unauthorized(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Request without authorization header
	req := httptest.NewRequest("GET", "/api/v1/available-locations", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Missing authorization header")
}

func TestGetAvailableLocations_InvalidToken(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Request with invalid token
	req := httptest.NewRequest("GET", "/api/v1/available-locations", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid or expired token")
}

func TestGetAvailableLocations_RegularAdminAccess(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create regular admin (not super admin)
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "regularadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&admin)

	token, _ := utils.GenerateAdminToken(admin.ID, admin.Username, admin.Role, 0)

	req := httptest.NewRequest("GET", "/api/v1/available-locations", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Regular admins should also be able to access available locations
	assert.NotEqual(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, fiber.StatusForbidden, resp.StatusCode)

	var response AvailableLocationsResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Should have proper response structure
	assert.NotNil(t, response.Data)
}
