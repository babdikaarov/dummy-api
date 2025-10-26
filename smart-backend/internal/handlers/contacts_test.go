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

func TestGetContact_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create contact information
	contact := models.Contact{
		SupportNumber: 77091234567,
		EmailSupport:  "support@ololo.com",
		Address:       "г. Бишкек, проспект Чуй, 135",
	}
	db.DB.Create(&contact)

	req := httptest.NewRequest("GET", "/api/v1/contacts", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ContactResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Contact information retrieved successfully", response.Message)
	assert.Equal(t, 77091234567, response.Data.SupportNumber)
	assert.Equal(t, "support@ololo.com", response.Data.EmailSupport)
	assert.Equal(t, "г. Бишкек, проспект Чуй, 135", response.Data.Address)
}

func TestGetContact_NoContactInfo(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Don't create any contact info - database starts empty

	req := httptest.NewRequest("GET", "/api/v1/contacts", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ContactResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	// Should return empty/default values
	assert.Equal(t, 0, response.Data.SupportNumber)
	assert.Equal(t, "", response.Data.EmailSupport)
	assert.Equal(t, "", response.Data.Address)
}

func TestUpdateContact_CreateNew(t *testing.T) {
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

	// Create new contact
	updateReq := UpdateContactRequest{
		SupportNumber: 77091234567,
		EmailSupport:  "support@ololo.com",
		Address:       "г. Бишкек, проспект Чуй, 135",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PATCH", "/api/v1/contacts", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ContactResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "successfully")
	assert.Equal(t, 77091234567, response.Data.SupportNumber)
	assert.Equal(t, "support@ololo.com", response.Data.EmailSupport)

	// Verify in database
	var savedContact models.Contact
	db.DB.First(&savedContact)
	assert.Equal(t, 77091234567, savedContact.SupportNumber)
}

func TestUpdateContact_UpdateExisting(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create existing contact
	contact := models.Contact{
		SupportNumber: 77011111111,
		EmailSupport:  "old@ololo.com",
		Address:       "Old Address",
	}
	db.DB.Create(&contact)

	// Create admin
	admin := models.Admin{
		ID:       uuid.New(),
		Username: "admin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&admin)

	token, _ := utils.GenerateAdminToken(admin.ID, admin.Username, admin.Role, 0)

	// Update contact
	updateReq := UpdateContactRequest{
		SupportNumber: 77099999999,
		EmailSupport:  "new@ololo.com",
		Address:       "New Address",
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PATCH", "/api/v1/contacts", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response ContactResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, 77099999999, response.Data.SupportNumber)
	assert.Equal(t, "new@ololo.com", response.Data.EmailSupport)
	assert.Equal(t, "New Address", response.Data.Address)

	// Verify in database - should update, not create new
	var allContacts []models.Contact
	db.DB.Find(&allContacts)
	assert.Equal(t, 1, len(allContacts)) // Only 1 contact should exist
	assert.Equal(t, 77099999999, allContacts[0].SupportNumber)
}

func TestUpdateContact_Unauthorized(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	updateReq := UpdateContactRequest{
		SupportNumber: 77091234567,
		EmailSupport:  "support@ololo.com",
		Address:       "г. Бишкек, проспект Чуй, 135",
	}
	reqBody, _ := json.Marshal(updateReq)

	// Request without authorization header
	req := httptest.NewRequest("PATCH", "/api/v1/contacts", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Missing authorization header")
}

func TestUpdateContact_InvalidJSON(t *testing.T) {
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

	// Send invalid JSON
	req := httptest.NewRequest("PATCH", "/api/v1/contacts", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid request body")
}
