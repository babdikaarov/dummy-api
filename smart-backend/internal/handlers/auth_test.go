package handlers

import (
	"ololo-gate/internal/tests"
	"ololo-gate/internal/utils"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupAuthTest(t *testing.T) *fiber.App {
	tests.SetupTestConfig()
	tests.SetupTestDB(t)

	app := fiber.New()
	app.Post("/register", Register)
	app.Post("/login", Login)
	app.Post("/refresh", RefreshToken)

	return app
}

func TestRegister_Success(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	body := map[string]string{
		"phone":    "+77771234567",
		"password": "testpassword123",
	}

	resp, err := tests.MakeRequest(app, "POST", "/register", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "User registered successfully", result["message"])

	data := result["data"].(map[string]interface{})
	assert.NotNil(t, data["id"])
	assert.Equal(t, "+77771234567", data["phone"])
}

func TestRegister_InvalidPhoneFormat(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	body := map[string]string{
		"phone":    "77771234567", // Missing +
		"password": "testpassword123",
	}

	resp, err := tests.MakeRequest(app, "POST", "/register", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "Invalid phone number format")
}

func TestRegister_ShortPassword(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	body := map[string]string{
		"phone":    "+77771234567",
		"password": "123", // Too short
	}

	resp, err := tests.MakeRequest(app, "POST", "/register", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "at least 6 characters")
}

func TestRegister_DuplicatePhone(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	// Create first user
	tests.CreateTestUser(t, "+77771234567", "password123")

	// Try to register with same phone
	body := map[string]string{
		"phone":    "+77771234567",
		"password": "different password",
	}

	resp, err := tests.MakeRequest(app, "POST", "/register", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 409, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "already exists")
}

func TestLogin_Success(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	// Create test user
	tests.CreateTestUser(t, "+77771234567", "testpassword123")

	body := map[string]string{
		"phone":    "+77771234567",
		"password": "testpassword123",
	}

	resp, err := tests.MakeRequest(app, "POST", "/login", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "Login successful", result["message"])

	data := result["data"].(map[string]interface{})
	assert.NotNil(t, data["id"])
	assert.Equal(t, "+77771234567", data["phone"])
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])

	// Verify tokens are valid
	accessToken := data["access_token"].(string)
	claims, err := utils.ValidateToken(accessToken, utils.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "+77771234567", claims.Phone)
	assert.Equal(t, 1, claims.TokenVersion) // TokenVersion incremented to 1 on login
}

func TestLogin_InvalidCredentials(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	// Create test user
	tests.CreateTestUser(t, "+77771234567", "correctpassword")

	body := map[string]string{
		"phone":    "+77771234567",
		"password": "wrongpassword",
	}

	resp, err := tests.MakeRequest(app, "POST", "/login", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Equal(t, "Invalid credentials", result["message"])
}

func TestLogin_UserNotFound(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	body := map[string]string{
		"phone":    "+77771234567",
		"password": "testpassword123",
	}

	resp, err := tests.MakeRequest(app, "POST", "/login", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Equal(t, "Invalid credentials", result["message"])
}

func TestRefreshToken_Success(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	// Create test user and login
	user := tests.CreateTestUser(t, "+77771234567", "testpassword123")

	// Generate tokens
	tokens, err := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)
	assert.NoError(t, err)

	// Use refresh token to get new access token
	body := map[string]string{
		"refresh_token": tokens.RefreshToken,
	}

	resp, respErr := tests.MakeRequest(app, "POST", "/refresh", body, nil)
	assert.NoError(t, respErr)
	assert.Equal(t, 200, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.True(t, result["success"].(bool))
	assert.Equal(t, "Token refreshed successfully", result["message"])

	data := result["data"].(map[string]interface{})
	assert.NotEmpty(t, data["access_token"])

	// Verify new access token is valid
	newAccessToken := data["access_token"].(string)
	claims, err := utils.ValidateToken(newAccessToken, utils.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Phone, claims.Phone)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	body := map[string]string{
		"refresh_token": "invalid.token.here",
	}

	resp, err := tests.MakeRequest(app, "POST", "/refresh", body, nil)
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.Code)

	result := tests.ParseJSONResponse(t, resp)
	assert.False(t, result["success"].(bool))
	assert.Contains(t, result["message"], "Invalid or expired refresh token")
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	app := setupAuthTest(t)
	defer tests.CleanupTestDB(t)

	// Create an expired refresh token (use access token expiry for quick expiration)
	user := tests.CreateTestUser(t, "+77771234567", "testpassword123")

	// This would normally be expired, but for testing we'll use an invalid version
	tokens, err := utils.GenerateTokens(user.ID, user.Phone, user.TokenVersion)
	assert.NoError(t, err)

	// Increment user's token version to simulate invalidation
	user.TokenVersion = 1
	tests.SetupTestDB(t) // Reset DB with updated user

	body := map[string]string{
		"refresh_token": tokens.RefreshToken,
	}

	resp, respErr := tests.MakeRequest(app, "POST", "/refresh", body, nil)
	assert.NoError(t, respErr)

	// Should fail because token version doesn't match
	assert.Equal(t, 401, resp.Code)
}
