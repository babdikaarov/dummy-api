package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"ololo-gate/internal/db"
	"ololo-gate/internal/models"
	"ololo-gate/internal/utils"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetAllAdmins_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	// Create regular admin
	regularAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "regularadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&regularAdmin)

	// Generate token
	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response AdminsListResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Admins retrieved successfully", response.Message)
	assert.GreaterOrEqual(t, len(response.Data), 2)
	assert.GreaterOrEqual(t, response.Pagination.Total, 2)
}

func TestGetAllAdmins_Unauthorized(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestGetAllAdmins_RegularAdminForbidden(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create regular admin
	regularAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "regularadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&regularAdmin)

	// Generate token for regular admin
	token, _ := utils.GenerateAdminToken(regularAdmin.ID, regularAdmin.Username, regularAdmin.Role, 0)

	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response["success"].(bool))
	assert.Equal(t, "Super admin access required", response["message"])
}

func TestCreateAdmin_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	// Create new admin request
	createReq := CreateAdminRequest{
		Username: "newadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Admin created successfully", response.Message)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, "newadmin", data["username"])
	assert.Equal(t, models.RoleRegular, data["role"])
}

func TestCreateAdmin_InvalidRole(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	// Create admin with invalid role
	createReq := CreateAdminRequest{
		Username: "newadmin",
		Password: "password123",
		Role:     "invalid",
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Invalid role. Must be 'super' or 'regular'", response.Message)
}

func TestCreateAdmin_ShortPassword(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	// Create admin with short password
	createReq := CreateAdminRequest{
		Username: "newadmin",
		Password: "12345",
		Role:     models.RoleRegular,
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Password must be at least 6 characters long", response.Message)
}

func TestCreateAdmin_DuplicateUsername(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	// Create existing admin
	existingAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "existing",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&existingAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	// Try to create admin with duplicate username
	createReq := CreateAdminRequest{
		Username: "existing",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	reqBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/v1/admin/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Admin with this username already exists", response.Message)
}

func TestUpdateAdminPassword_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	// Create admin to update
	targetAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "targetadmin",
		Password: "oldpassword",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&targetAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	// Update password request
	newPassword := "newpassword123"
	updateReq := UpdateAdminRequest{
		Password: &newPassword,
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/users/%s", targetAdmin.ID.String()), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Admin updated successfully", response.Message)

	// Verify password was updated
	var updatedAdmin models.Admin
	db.DB.First(&updatedAdmin, targetAdmin.ID)
	assert.True(t, updatedAdmin.CheckPassword("newpassword123"))
}

func TestUpdateAdminPassword_NotFound(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	// Update password for non-existent admin
	newPassword := "newpassword123"
	updateReq := UpdateAdminRequest{
		Password: &newPassword,
	}
	reqBody, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/users/%s", uuid.New().String()), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Admin not found", response.Message)
}

func TestDeleteAdmin_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	// Create admin to delete
	targetAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "targetadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&targetAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/admin/users/%s", targetAdmin.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Admin deleted successfully", response.Message)

	// Verify admin was soft deleted
	var deletedAdmin models.Admin
	result := db.DB.Unscoped().First(&deletedAdmin, targetAdmin.ID)
	assert.NoError(t, result.Error)
	assert.NotNil(t, deletedAdmin.DeletedAt)
}

func TestDeleteAdmin_NotFound(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/admin/users/%s", uuid.New().String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Admin not found", response.Message)
}

func TestGetAdminByID_Success(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	// Create regular admin
	regularAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "regularadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&regularAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/admin/users/%s", regularAdmin.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response AdminDetailResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, "Admin retrieved successfully", response.Message)
	assert.Equal(t, regularAdmin.ID.String(), response.Data.AdminID.String())
	assert.Equal(t, "regularadmin", response.Data.Username)
	assert.Equal(t, models.RoleRegular, response.Data.Role)
}

func TestGetAdminByID_InvalidIDFormat(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	req := httptest.NewRequest("GET", "/api/v1/admin/users/invalid-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid admin ID format")
}

func TestGetAdminByID_NotFound(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create super admin
	superAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "superadmin",
		Password: "password123",
		Role:     models.RoleSuper,
	}
	db.DB.Create(&superAdmin)

	token, _ := utils.GenerateAdminToken(superAdmin.ID, superAdmin.Username, superAdmin.Role, 0)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/admin/users/%s", uuid.New().String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.False(t, response.Success)
	assert.Equal(t, "Admin not found", response.Message)
}

func TestGetAdminByID_RegularAdminOwnAccess(t *testing.T) {
	app, cleanup := SetupTestApp()
	defer cleanup()

	// Create regular admin
	regularAdmin := models.Admin{
		ID:       uuid.New(),
		Username: "regularadmin",
		Password: "password123",
		Role:     models.RoleRegular,
	}
	db.DB.Create(&regularAdmin)

	token, _ := utils.GenerateAdminToken(regularAdmin.ID, regularAdmin.Username, regularAdmin.Role, 0)

	// Regular admin should be able to access their own profile
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/admin/users/%s", regularAdmin.ID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var response AdminDetailResponse
	json.NewDecoder(resp.Body).Decode(&response)

	assert.True(t, response.Success)
	assert.Equal(t, regularAdmin.ID.String(), response.Data.AdminID.String())
}
