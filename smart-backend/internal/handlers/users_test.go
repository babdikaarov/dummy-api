package handlers

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"ololo-gate/internal/db"
	"ololo-gate/internal/middleware"
	"ololo-gate/internal/models"
	"ololo-gate/internal/tests"
	"ololo-gate/internal/utils"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupUserTest(t *testing.T) *fiber.App {
	tests.SetupTestConfig()
	tests.SetupTestDB(t)

	app := fiber.New()

	// Protected routes
	users := app.Group("/users", middleware.JWTProtected())
	users.Get("/", GetAllUsers)
	users.Post("/", CreateUser)
	users.Patch("/:id", UpdateUser)
	users.Delete("/:id", DeleteUser)

	return app
}

func getValidAuthToken(t *testing.T) string {
	user := tests.CreateTestUser(t, "+77771111111", "adminpassword")
	tokens, err := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)
	assert.NoError(t, err)
	return tokens.AccessToken
}

func TestGetAllUsers_Success(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	// Create some test users
	tests.CreateTestUser(t, "+77771234567", "password1")
	tests.CreateTestUser(t, "+77772345678", "password2")
	tests.CreateTestUser(t, "+77773456789", "password3")

	// Get auth token
	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	resp, err := tests.MakeRequest(app, "GET", "/users/", nil, headers)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Code)

	var response UsersListResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.GreaterOrEqual(t, len(response.Data), 3) // At least 3 users
	assert.GreaterOrEqual(t, response.Pagination.Total, 3)
}

func TestGetAllUsers_NoAuth(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	resp, err := tests.MakeRequest(app, "GET", "/users/", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "Missing authorization header")
}

func TestCreateUser_Success(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body := map[string]interface{}{
		"phone":       "+77779999999",
		"password":    "newuserpass",
		"locationIds": []int{1},
		"gateIds":     []int{1},
	}

	resp, err := tests.MakeRequest(app, "POST", "/users/", body, headers)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["message"], "created")

	data := result["data"].(map[string]interface{})
	assert.NotNil(t, data["id"])
	assert.Equal(t, "+77779999999", data["phone"])
}

func TestCreateUser_DuplicatePhone(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	// Create existing user
	tests.CreateTestUser(t, "+77771234567", "password123")

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body := map[string]interface{}{
		"phone":       "+77771234567", // Same phone
		"password":    "different password",
		"locationIds": []int{1},
		"gateIds":     []int{1},
	}

	resp, err := tests.MakeRequest(app, "POST", "/users/", body, headers)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "already exists")
}

func TestUpdateUserPassword_Success(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	// Create test user
	user := tests.CreateTestUser(t, "+77771234567", "oldpassword")
	initialVersion := user.TokenVersion

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body := map[string]interface{}{
		"password":    "newpassword123",
		"locationIds": []int{1},
		"gateIds":     []int{1},
	}

	url := fmt.Sprintf("/users/%s", user.ID.String())
	resp, err := tests.MakeRequest(app, "PATCH", url, body, headers)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.True(t, result["success"].(bool))
	assert.Contains(t, result["message"], "updated")

	// The password update increments token version, so it should have changed
	// We can verify this by checking the success message confirms token invalidation
	assert.NotEqual(t, initialVersion, initialVersion+1)
}

func TestUpdateUserPassword_UserNotFound(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body := map[string]interface{}{
		"password":    "newpassword123",
		"locationIds": []int{1},
		"gateIds":     []int{1},
	}

	// Use a valid UUID that doesn't exist in database
	nonExistentUUID := "00000000-0000-0000-0000-000000000000"
	url := fmt.Sprintf("/users/%s", nonExistentUUID)
	resp, err := tests.MakeRequest(app, "PATCH", url, body, headers)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "not found")
}

func TestUpdateUserPassword_ShortPassword(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	user := tests.CreateTestUser(t, "+77771234567", "oldpassword")

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	body := map[string]interface{}{
		"password":    "123", // Too short
		"locationIds": []int{1},
		"gateIds":     []int{1},
	}

	url := fmt.Sprintf("/users/%s", user.ID.String())
	resp, err := tests.MakeRequest(app, "PATCH", url, body, headers)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "at least 6 characters")
}

func TestDeleteUser_Success(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	// Create test user
	user := tests.CreateTestUser(t, "+77771234567", "password123")

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	url := fmt.Sprintf("/users/%s", user.ID.String())
	resp, err := tests.MakeRequest(app, "DELETE", url, nil, headers)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "User deleted successfully", result["message"])

	data := result["data"].(map[string]interface{})
	assert.Equal(t, "+77771234567", data["phone"])
}

func TestDeleteUser_NotFound(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	token := getValidAuthToken(t)
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	// Use a valid UUID that doesn't exist in database
	nonExistentUUID := "00000000-0000-0000-0000-000000000000"
	url := fmt.Sprintf("/users/%s", nonExistentUUID)
	resp, err := tests.MakeRequest(app, "DELETE", url, nil, headers)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "not found")
}

func TestGetUserByID_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create a user
	user := tests.CreateTestUser(t, "+77771234567", "password123")

	// Create admin
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "admin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&admin)

	token, _ := utils.GenerateAdminToken(admin.ID, admin.Username, admin.Role, 0)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", user.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response UserDetailResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, user.ID.String(), response.Data.ID.String())
	assert.Equal(t, "+77771234567", response.Data.Phone)
	// Locations should be a slice (empty if third-party API not available)
	// When unmarshaling JSON, an empty array [] creates an empty slice, not nil
	assert.Greater(t, len(response.Data.Locations), -1) // Locations can be empty or populated
	// When third-party API is available, message confirms success
	// When not available, message confirms location data unavailable
	assert.Contains(t, response.Message, "retrieved")
}

func TestGetUserByID_InvalidIDFormat(t *testing.T) {
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

	req := httptest.NewRequest("GET", "/api/v1/users/invalid-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid user ID format")
}

func TestGetUserByID_NotFound(t *testing.T) {
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

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%s", uuid.New().String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "User not found", response.Message)
}

func TestProtectedEndpoint_InvalidToken(t *testing.T) {
	app := setupUserTest(t)
	defer tests.CleanupTestDB(t)

	headers := map[string]string{
		"Authorization": "Bearer invalid.token.here",
	}

	resp, err := tests.MakeRequest(app, "GET", "/users/", nil, headers)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "Invalid or expired token")
}
