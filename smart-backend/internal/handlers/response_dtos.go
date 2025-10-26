package handlers

import (
	"time"

	"github.com/google/uuid"
)

// Response DTOs for Swagger documentation and TypeScript generation
// These structs define the exact structure of successful API responses

// ========== Health Check Response ==========

// HealthCheckResponse defines the response structure for the health check endpoint
// @name HealthCheckResponse
type HealthCheckResponse struct {
	Success     bool   `json:"success" example:"true" validate:"required"`
	Message     string `json:"message" example:"Ololo Gate API is running" validate:"required"`
	Status      string `json:"status" example:"healthy" validate:"required"`
	Timestamp   string `json:"timestamp" example:"2025-01-15T10:30:45Z" validate:"required"`
	Uptime      string `json:"uptime" example:"1h30m45s" validate:"required"`
	Environment string `json:"environment" example:"production" validate:"required"`
	Version     string `json:"version" example:"1.0.0" validate:"required"`
}

// ========== Pagination ==========

// PaginationMeta defines the pagination metadata for list responses
// @name PaginationMeta
type PaginationMeta struct {
	Total       int `json:"total" example:"100"`
	PerPage     int `json:"per_page" example:"100"`
	CurrentPage int `json:"current_page" example:"1"`
	LastPage    int `json:"last_page" example:"1"`
}

// ========== User Authentication Responses ==========

// RegisterResponse defines the response structure for successful user registration
// @name RegisterResponse
type RegisterResponse struct {
	Success bool         `json:"success" example:"true" validate:"required"`
	Message string       `json:"message" example:"User registered successfully" validate:"required"`
	Data    RegisterData `json:"data"`
}

// @name RegisterData
type RegisterData struct {
	UserID uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Phone  string    `json:"phone" example:"+77771234567" validate:"required"`
}

// LoginResponse defines the response structure for successful user login
// @name LoginResponse
type LoginResponse struct {
	Success bool      `json:"success" example:"true" validate:"required"`
	Message string    `json:"message" example:"Login successful" validate:"required"`
	Data    LoginData `json:"data"`
}

// @name LoginData
type LoginData struct {
	UserID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Phone            string    `json:"phone" example:"+77771234567" validate:"required"`
	AccessToken      string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
	RefreshToken     string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
	AccessExpiresIn  int64     `json:"access_expires_in" example:"900" validate:"required"`
	RefreshExpiresIn int64     `json:"refresh_expires_in" example:"2592000" validate:"required"`
}

// RefreshResponse defines the response structure for successful token refresh
// @name RefreshResponse
type RefreshResponse struct {
	Success bool        `json:"success" example:"true" validate:"required"`
	Message string      `json:"message" example:"Token refreshed successfully" validate:"required"`
	Data    RefreshData `json:"data"`
}

// @name RefreshData
type RefreshData struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
}

// PhoneAvailabilityResponse defines the response structure for phone number availability check
// @name PhoneAvailabilityResponse
type PhoneAvailabilityResponse struct {
	Success   bool   `json:"success" example:"true" validate:"required"`
	Message   string `json:"message" example:"Phone availability checked" validate:"required"`
	Available bool   `json:"available" example:"true" validate:"required"` // true if phone is available, false if already in use
}

// ========== User Management Responses ==========

// UsersListResponse defines the response structure for retrieving all users with pagination
// @name UsersListResponse
type UsersListResponse struct {
	Success    bool             `json:"success" example:"true" validate:"required"`
	Message    string           `json:"message" example:"Users retrieved successfully" validate:"required"`
	Data       []UserDTO        `json:"data"`
	Pagination PaginationMeta   `json:"pagination"`
}

// @name UsersListData
type UsersListData struct {
	Users []UserDTO `json:"users" validate:"required"`
	Count int       `json:"count" example:"10" validate:"required"`
}

// @name UserDTO
type UserDTO struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Phone     string    `json:"phone" example:"+77771234567" validate:"required"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T10:30:00Z" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T10:30:00Z" validate:"required"`
}

// UserDetailDTO includes user info plus their assigned locations/gates
// @name UserDetailDTO
type UserDetailDTO struct {
	ID        uuid.UUID     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Phone     string        `json:"phone" example:"+77771234567" validate:"required"`
	CreatedAt time.Time     `json:"created_at" example:"2025-01-15T10:30:00Z" validate:"required"`
	UpdatedAt time.Time     `json:"updated_at" example:"2025-01-15T10:30:00Z" validate:"required"`
	Locations []LocationDTO `json:"locations" validate:"required"`
}

// UserResponse defines the response structure for user operations (create, update, delete)
// @name UserResponse
type UserResponse struct {
	Success bool     `json:"success" example:"true" validate:"required"`
	Message string   `json:"message" example:"User created successfully" validate:"required"`
	Data    UserData `json:"data"`
}

// UserDetailResponse defines the response structure for retrieving user details
// @name UserDetailResponse
type UserDetailResponse struct {
	Success bool          `json:"success" example:"true" validate:"required"`
	Message string        `json:"message" example:"User retrieved successfully" validate:"required"`
	Data    UserDetailDTO `json:"data"`
}

// @name UserData
type UserData struct {
	UserID uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`
	Phone  string    `json:"phone" example:"+77771234567" validate:"required"`
}

// ========== Admin Authentication Responses ==========

// AdminLoginResponse defines the response structure for successful admin login
// @name AdminLoginResponse
type AdminLoginResponse struct {
	Success bool           `json:"success" example:"true" validate:"required"`
	Message string         `json:"message" example:"Login successful" validate:"required"`
	Data    AdminLoginData `json:"data"`
}

// @name AdminLoginData
type AdminLoginData struct {
	AdminID     uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000001" validate:"required"`
	Username    string    `json:"username" example:"admin" validate:"required"`
	Role        string    `json:"role" example:"super" validate:"required"`
	AccessToken string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"`
}

// ========== Admin Management Responses ==========

// AdminsListResponse defines the response structure for retrieving all admins with pagination
// @name AdminsListResponse
type AdminsListResponse struct {
	Success    bool             `json:"success" example:"true" validate:"required"`
	Message    string           `json:"message" example:"Admins retrieved successfully" validate:"required"`
	Data       []AdminDTO       `json:"data"`
	Pagination PaginationMeta   `json:"pagination"`
}

// @name AdminsListData
type AdminsListData struct {
	Admins []AdminDTO `json:"admins" validate:"required"`
	Count  int        `json:"count" example:"5" validate:"required"`
}

// @name AdminDTO
type AdminDTO struct {
	ID        uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000001" validate:"required"`
	Username  string    `json:"username" example:"admin" validate:"required"`
	Role      string    `json:"role" example:"super" validate:"required"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T10:30:00Z" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T10:30:00Z" validate:"required"`
}

// AdminResponse defines the response structure for admin operations (create, update, delete)
// @name AdminResponse
type AdminResponse struct {
	Success bool      `json:"success" example:"true" validate:"required"`
	Message string    `json:"message" example:"Admin created successfully" validate:"required"`
	Data    AdminData `json:"data"`
}

// @name AdminData
type AdminData struct {
	AdminID  uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440001" validate:"required"`
	Username string    `json:"username" example:"newadmin" validate:"required"`
	Role     string    `json:"role" example:"regular" validate:"required"`
}

// AdminDetailResponse defines the response structure for retrieving admin details by ID
// @name AdminDetailResponse
type AdminDetailResponse struct {
	Success bool               `json:"success" example:"true"`
	Message string             `json:"message" example:"Admin retrieved successfully"`
	Data    AdminDetailData    `json:"data"`
}

// @name AdminDetailData
type AdminDetailData struct {
	AdminID   uuid.UUID `json:"id" example:"00000000-0000-0000-0000-000000000001"`
	Username  string    `json:"username" example:"admin"`
	Role      string    `json:"role" example:"super"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2025-01-15T10:30:00Z"`
}

// ========== Gate Management Responses ==========

// GateDTO represents a single gate/barrier
// @name GateDTO
type GateDTO struct {
	ID               int    `json:"id" example:"1"`
	Title            string `json:"title" example:"Автоматический Шлагбаум №12"`
	Description      string `json:"description" example:"Main vehicle entrance for visitors. Controlled by biometric access, opens in 3 seconds with safety sensors."`
	LocationID       int    `json:"location_id" example:"1"`
	IsOpen           bool   `json:"is_open" example:"true"`
	GateIsHorizontal bool   `json:"gate_is_horizontal" example:"true"`
}

// LocationDTO represents a location/facility with associated gates
// @name LocationDTO
type LocationDTO struct {
	ID      int       `json:"id" example:"1"`
	Title   string    `json:"title" example:"Торгово-развлекательный центр Ала-Тоо"`
	Address string    `json:"address" example:"г. Бишкек, проспект Чуй, 135"`
	Logo    string    `json:"logo" example:"https://picsum.photos/seed/alatoo/200"`
	Gates   []GateDTO `json:"gates"` // Always include gates, even if empty array
}

// LocationsListResponse defines the response structure for retrieving all locations
// @name LocationsListResponse
type LocationsListResponse struct {
	Success bool          `json:"success" example:"true" validate:"required"`
	Message string        `json:"message" example:"Locations retrieved successfully" validate:"required"`
	Data    []LocationDTO `json:"data"`
}

// GatesListResponse defines the response structure for retrieving gates for a location
// @name GatesListResponse
type GatesListResponse struct {
	Success bool      `json:"success" example:"true" validate:"required"`
	Message string    `json:"message" example:"Gates retrieved successfully" validate:"required"`
	Data    []GateDTO `json:"data"`
}

// GateActionData represents the response data for gate open/close operations
// @name GateActionData
type GateActionData struct {
	GateID int  `json:"gate_id" example:"1"`
	Status bool `json:"status" example:"true"`
}

// GateActionResponse defines the response structure for gate operations (open/close)
// @name GateActionResponse
type GateActionResponse struct {
	Success bool            `json:"success" example:"true" validate:"required"`
	Message string          `json:"message" example:"Gate operation completed successfully" validate:"required"`
	Data    GateActionData  `json:"data"`
}

// ========== Contact Information Responses ==========

// ContactDTO represents the contact information
// @name ContactDTO
type ContactDTO struct {
	SupportNumber int       `json:"support_number" example:"77091234567"`
	EmailSupport  string    `json:"email_support" example:"support@ololo.com"`
	Address       string    `json:"address" example:"г. Бишкек, проспект Чуй, 135"`
}

// ContactResponse defines the response structure for contact information
// @name ContactResponse
type ContactResponse struct {
	Success bool       `json:"success" example:"true" validate:"required"`
	Message string     `json:"message" example:"Contact information retrieved successfully" validate:"required"`
	Data    ContactDTO `json:"data"`
}

// ========== User Creation/Update with Location Assignment ==========

// LocationAssignmentRequest represents a location with its assigned gates
// @name LocationAssignmentRequest
type LocationAssignmentRequest struct {
	LocationID int   `json:"locationId" example:"1" validate:"required"`
	GateIds    []int `json:"gateIds" validate:"required"`
}

// CreateUserRequest defines the structure for creating a new user with optional location/gate assignment
// @name CreateUserRequest
type CreateUserRequest struct {
	Phone     string                        `json:"phone" example:"+77771234567" validate:"required"`
	Password  string                        `json:"password" example:"password123" validate:"required,min=6"`
	Locations []LocationAssignmentRequest   `json:"locations"` // Optional - if provided, will assign user to these locations and gates
}

// UpdateUserRequest defines the structure for updating a user (all fields optional)
// @name UpdateUserRequest
type UpdateUserRequest struct {
	Phone     string                        `json:"phone" example:"+77771234567"` // Optional - if provided, will update phone number after checking availability
	Password  string                        `json:"password" example:"newpassword123" validate:"omitempty,min=6"` // Optional - only updates if provided
	Locations []LocationAssignmentRequest   `json:"locations"` // Optional - if provided, will reassign user to these locations and gates
}

// ========== Available Locations Response ==========

// AvailableLocationsResponse defines the response for all available locations
// @name AvailableLocationsResponse
type AvailableLocationsResponse struct {
	Success bool           `json:"success" example:"true" validate:"required"`
	Message string         `json:"message" example:"Available locations retrieved successfully" validate:"required"`
	Data    []LocationDTO  `json:"data"`
}
