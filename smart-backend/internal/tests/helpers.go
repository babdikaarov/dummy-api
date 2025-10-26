package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"ololo-gate/internal/config"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) {
	var err error
	db.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate test models
	err = db.DB.AutoMigrate(&models.User{}, &models.Admin{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
}

// SetupTestConfig initializes test configuration
func SetupTestConfig() {
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key",
			AccessExpiry:  900000000000,   // 15 minutes in nanoseconds
			RefreshExpiry: 2592000000000000, // 30 days in nanoseconds
		},
		Server: config.ServerConfig{
			Port: "8080",
			Env:  "test",
		},
	}
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, phone, password string) *models.User {
	user := &models.User{
		Phone:        phone,
		Password:     password,
		TokenVersion: 0,
	}

	if err := db.DB.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// MakeRequest helper function to make HTTP requests in tests
func MakeRequest(app *fiber.App, method, url string, body interface{}, headers map[string]string) (*httptest.ResponseRecorder, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		return nil, err
	}

	recorder := httptest.NewRecorder()
	recorder.WriteHeader(resp.StatusCode)
	io.Copy(recorder, resp.Body)

	return recorder, nil
}

// ParseJSONResponse parses JSON response body into a map
func ParseJSONResponse(t *testing.T, resp *httptest.ResponseRecorder) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	return result
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T) {
	// Delete all users and admins
	if err := db.DB.Exec("DELETE FROM users").Error; err != nil {
		t.Logf("Warning: Failed to cleanup users: %v", err)
	}
	if err := db.DB.Exec("DELETE FROM admins").Error; err != nil {
		t.Logf("Warning: Failed to cleanup admins: %v", err)
	}
}
