package handlers

import (
	"bytes"
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

func TestAdminLogin_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create test admin
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "testadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&admin)

	// Login request
	loginReq := AdminLoginRequest{
		Username: "testadmin",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Login successful", response.Message)
	assert.NotNil(t, response.Data)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, admin.ID.String(), data["id"])
	assert.Equal(t, "testadmin", data["username"])
	assert.Equal(t, models.RoleSuper, data["role"])
	assert.NotEmpty(t, data["access_token"])

	// Verify token is valid and permanent (no expiry)
	token := data["access_token"].(string)
	claims, err := utils.ValidateAdminToken(token)
	assert.NoError(t, err)
	assert.Equal(t, admin.ID, claims.AdminID)
	assert.Equal(t, "testadmin", claims.Username)
	assert.Equal(t, models.RoleSuper, claims.Role)
	assert.Nil(t, claims.ExpiresAt) // Permanent token has no expiry
}

func TestAdminLogin_InvalidUsername(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	loginReq := AdminLoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid credentials", response.Message)
}

func TestAdminLogin_InvalidPassword(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create test admin
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "testadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&admin)

	// Login with wrong password
	loginReq := AdminLoginRequest{
		Username: "testadmin",
		Password: "wrongpassword",
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid credentials", response.Message)
}

func TestAdminLogin_MissingUsername(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	loginReq := AdminLoginRequest{
		Username: "",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Username and password are required", response.Message)
}

func TestAdminLogin_MissingPassword(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	loginReq := AdminLoginRequest{
		Username: "testadmin",
		Password: "",
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Username and password are required", response.Message)
}

func TestAdminLogin_InvalidJSON(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request body", response.Message)
}

func TestAdminLogin_RegularAdminRole(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create regular admin
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "regularadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&admin)

	// Login request
	loginReq := AdminLoginRequest{
		Username: "regularadmin",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	data := response.Data.(map[string]interface{})
	assert.Equal(t, models.RoleRegular, data["role"])
}
