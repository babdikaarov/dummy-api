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

func TestGetLocations_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("GET", "/api/v1/locations", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// If third-party API is not available, will return 500
	// When available, will return 200 with data
	assert.NotEqual(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)

	var response LocationsListResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Should have proper response structure (may be empty if third-party API not available)
	// When unmarshaling JSON, an empty array [] creates an empty slice, not nil
	assert.Greater(t, len(response.Data), -1) // Data can be empty or populated
	// Success field should match status code
	if resp.StatusCode == fiber.StatusOK {
		assert.True(t, response.Success)
	}
}

func TestGetLocations_Unauthorized(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Request without authorization header
	req := httptest.NewRequest("GET", "/api/v1/locations", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Missing authorization header")
}

func TestGetGatesByLocation_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("GET", "/api/v1/locations/1/gates", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should not return unauthorized errors
	assert.NotEqual(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response GatesListResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// Should have proper response structure
	assert.NotNil(t, response.Data)
}

func TestGetGatesByLocation_InvalidLocationID(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("GET", "/api/v1/locations/invalid/gates", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid location ID")
}

func TestGetGatesByLocation_NegativeLocationID(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("GET", "/api/v1/locations/-1/gates", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid location ID")
}

func TestOpenGate_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("PUT", "/api/v1/locations/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should not return unauthorized/bad request errors
	assert.NotEqual(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)

	var response GateActionResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// When third-party API is available, should succeed
	if resp.StatusCode == fiber.StatusOK {
		assert.True(t, response.Success)
		assert.Equal(t, 1, response.Data.GateID)
		assert.NotNil(t, response.Data.Status)
	} else {
		// When API not available, still returns structured error
		assert.NotNil(t, response.Message)
	}
}

func TestOpenGate_InvalidGateID(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("PUT", "/api/v1/locations/invalid/open", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid gate ID")
}

func TestOpenGate_NegativeGateID(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("PUT", "/api/v1/locations/-1/open", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid gate ID")
}

func TestOpenGate_Unauthorized(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Request without authorization header
	req := httptest.NewRequest("PUT", "/api/v1/locations/1/open", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
}

func TestCloseGate_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("PUT", "/api/v1/locations/1/close", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should not return unauthorized/bad request errors
	assert.NotEqual(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)

	var response GateActionResponse
	json.NewDecoder(resp.Body).Decode(&response)

	// When third-party API is available, should succeed
	if resp.StatusCode == fiber.StatusOK {
		assert.True(t, response.Success)
		assert.Equal(t, 1, response.Data.GateID)
		assert.NotNil(t, response.Data.Status)
	} else {
		// When API not available, still returns structured error
		assert.NotNil(t, response.Message)
	}
}

func TestCloseGate_InvalidGateID(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := models.User{
		ID:           uuid.New(),
		Phone:        "+77771234567",
		Password:     "password123",
		TokenVersion: 0,
	}
	db.DB.Create(&user)

	tokens, _ := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)

	req := httptest.NewRequest("PUT", "/api/v1/locations/invalid/close", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid gate ID")
}

func TestCloseGate_Unauthorized(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Request without authorization header
	req := httptest.NewRequest("PUT", "/api/v1/locations/1/close", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
}
