# Ololo Gate Backend - System Diagrams

This file contains comprehensive diagrams for the Ololo Gate backend system.
Copy and paste each section into [Eraser.io](https://app.eraser.io/) to visualize.

---

## 1. System Architecture Diagram

```
title Ololo Gate Backend Architecture

// External Components
Flutter App [icon: mobile]
PostgreSQL [icon: database, color: blue]

// Main Application
Fiber App [icon: server, color: green] {
  HTTP Router
  CORS Middleware
}

// Handler Layer
User Handlers [icon: file-code, color: orange] {
  Register
  Login
  Refresh Token
}

User Management [icon: users, color: orange] {
  Get All Users
  Create User
  Update Password
  Delete User
}

Admin Auth [icon: shield, color: red] {
  Admin Login
}

Admin Management [icon: shield-check, color: red] {
  Manage Admins (Super Only)
  Create Admin
  Update Admin
  Delete Admin
}

Admin User Management [icon: user-cog, color: red] {
  Admin Get Users
  Admin Create User
  Admin Update User
  Admin Delete User
}

// Middleware Layer
JWT Middleware [icon: lock, color: yellow] {
  Validate User Token
  Check Token Version
}

Admin Middleware [icon: shield, color: yellow] {
  Validate Admin Token
  Check Admin Role
  Super Admin Only
}

// Business Logic
JWT Utils [icon: key, color: purple] {
  Generate User Tokens
  Generate Admin Tokens
  Validate Tokens
  Token Types (Access/Refresh/Admin)
}

// Data Layer
User Model [icon: user, color: cyan] {
  UUID
  Phone
  Password (bcrypt)
  Token Version
  Timestamps
}

Admin Model [icon: shield, color: cyan] {
  UUID
  Username
  Password (bcrypt)
  Role (super/regular)
  Timestamps
}

Database Layer [icon: database, color: blue] {
  GORM ORM
  Connection Pool
  Auto Migration
  Initial Admin Seed
}

Config [icon: settings, color: gray] {
  Environment Variables
  JWT Settings
  Database Settings
  Admin Init Settings
}

// Connections - User Flow
Flutter App > Fiber App: HTTP Requests
Fiber App > User Handlers: Route /api/v1/auth/*
Fiber App > User Management: Route /api/v1/users/*
User Management > JWT Middleware: Protected Routes

// Connections - Admin Flow
Flutter App > Admin Auth: POST /api/v1/admin/login
Fiber App > Admin Auth: Route /api/v1/admin/login
Fiber App > Admin Management: Route /api/v1/admin/admins/*
Fiber App > Admin User Management: Route /api/v1/admin/users/*
Admin Management > Admin Middleware: Super Admin Only
Admin User Management > Admin Middleware: All Admins

// Middleware to Utils
JWT Middleware > JWT Utils: Validate User Token
Admin Middleware > JWT Utils: Validate Admin Token

// Handlers to Utils
User Handlers > JWT Utils: Generate Tokens
Admin Auth > JWT Utils: Generate Admin Token
User Management > User Model: CRUD Operations
Admin Management > Admin Model: CRUD Operations
Admin User Management > User Model: CRUD Operations

// Models to Database
User Model > Database Layer: ORM Operations
Admin Model > Database Layer: ORM Operations
Database Layer > PostgreSQL: SQL Queries

// Config connections
Config > Fiber App: Server Config
Config > Database Layer: DB Config
Config > JWT Utils: JWT Secret
Config > Database Layer: Initial Admin Config
```

---

## 2. Database Schema (ERD)

```
// Users Table
users [icon: user, color: blue] {
  id uuid PK "Primary Key (UUID)"
  phone varchar UK "E.164 format, unique"
  password varchar "bcrypt hashed"
  token_version integer "For token invalidation"
  created_at timestamp
  updated_at timestamp
  deleted_at timestamp "Soft delete"
}

// Admins Table
admins [icon: shield, color: red] {
  id uuid PK "Primary Key (UUID)"
  username varchar UK "Unique username"
  password varchar "bcrypt hashed"
  role varchar "super or regular"
  created_at timestamp
  updated_at timestamp
  deleted_at timestamp "Soft delete"
}

// Notes
note "Users authenticate with phone number" [color: blue]
note "Admins authenticate with username" [color: red]
note "Token version increments on password change" [color: orange]
note "Admin tokens never expire" [color: green]
```

---

## 3. User Authentication Flow (Sequence Diagram)

```
title User Authentication & Token Refresh Flow

Flutter App > API: POST /api/v1/auth/register
note over API: Validate phone format (E.164)
API > Database: Check if phone exists
Database > API: Phone not found
API > Database: Create user (bcrypt password)
Database > API: User created
API > Flutter App: 201 Created {id, phone}

// Login Flow
Flutter App > API: POST /api/v1/auth/login
API > Database: Find user by phone
Database > API: User found
note over API: Compare password (bcrypt)
API > JWT Utils: Generate tokens (token_version: 0)
JWT Utils > API: {access_token, refresh_token}
API > Flutter App: 200 OK {access_token, refresh_token, id}

// Store tokens
note over Flutter App: Store tokens in secure storage

// API Request with Token
Flutter App > API: GET /api/v1/users
note right of API: Authorization: Bearer {access_token}
API > JWT Middleware: Validate token
JWT Middleware > JWT Utils: Parse & verify token
JWT Utils > JWT Middleware: Valid token (id, token_version)
JWT Middleware > Database: Get user token_version
Database > JWT Middleware: token_version: 0
note over JWT Middleware: Compare versions (0 == 0) 
JWT Middleware > API: User authenticated
API > Flutter App: 200 OK {users: [...]}

// Token Refresh Flow
Flutter App > API: POST /api/v1/auth/refresh
note right of API: {refresh_token}
API > JWT Utils: Validate refresh token
JWT Utils > Database: Check user & token_version
Database > JWT Utils: User valid, version matches
JWT Utils > API: New access_token
API > Flutter App: 200 OK {access_token}
```

---

## 4. Admin Authentication Flow (Sequence Diagram)

```
title Admin Authentication & Authorization Flow

// Initial Admin Creation (on startup)
App Startup > Config: Load INIT_ADMIN settings
Config > Database: Check if admin UUID exists
Database > Config: Admin not found
Config > Database: Create super admin
note over Database: username: "admin", role: "super"
Database > Config: Admin created
Config > App: Ready

// Admin Login
Admin Client > API: POST /api/v1/admin/login
note right of API: {username: "admin", password: "admin"}
API > Database: Find admin by username
Database > API: Admin found (role: super)
note over API: Verify password (bcrypt)
API > JWT Utils: Generate permanent admin token
note over JWT Utils: NO EXPIRY DATE
JWT Utils > API: {access_token} (permanent)
API > Admin Client: 200 OK {id, username, role, access_token}

// Super Admin Access
Admin Client > API: GET /api/v1/admin/admins
note right of API: Authorization: Bearer {admin_token}
API > Admin Middleware: Validate admin token
Admin Middleware > JWT Utils: Parse & verify token
JWT Utils > Admin Middleware: Valid (id, role: super)
Admin Middleware > SuperAdminOnly: Check role
note over SuperAdminOnly: role == "super" 
SuperAdminOnly > API: Access granted
API > Database: Get all admins
Database > API: Admins list
API > Admin Client: 200 OK {admins: [...]}

// Regular Admin Access (User Management)
Regular Admin > API: GET /api/v1/admin/users
note right of API: Authorization: Bearer {admin_token}
API > Admin Middleware: Validate admin token
Admin Middleware > JWT Utils: Parse token
JWT Utils > Admin Middleware: Valid (role: regular)
note over Admin Middleware: Regular admin can access users
Admin Middleware > API: Access granted
API > Database: Get all users
Database > API: Users list
API > Regular Admin: 200 OK {users: [...]}

// Regular Admin Blocked from Admin Management
Regular Admin > API: GET /api/v1/admin/admins
API > Admin Middleware: Validate token
Admin Middleware > JWT Utils: Parse token
JWT Utils > Admin Middleware: Valid (role: regular)
Admin Middleware > SuperAdminOnly: Check role
note over SuperAdminOnly: role != "super" 
SuperAdminOnly > Regular Admin: 403 Forbidden
note left of Regular Admin: "Super admin access required"
```

---

## 5. Token Invalidation Flow (Sequence Diagram)

```
title User Token Invalidation on Password Change

// User has valid tokens
User App > API: Request with access_token
note over API: token_version = 0 (stored in token)
API > Database: Get user (token_version = 0)
note over API: Versions match 
API > User App: 200 OK

// Admin updates user password
Admin > API: PATCH /api/v1/admin/users/{id}
note right of API: {password: "new_password"}
API > Admin Middleware: Validate admin token
Admin Middleware > API: Authorized
API > Database: Find user by ID
Database > API: User found (token_version: 0)
note over API: Hash new password (bcrypt)
note over API: Increment token_version: 0 � 1
API > Database: UPDATE user SET password, token_version = 1
Database > API: User updated
API > Admin: 200 OK "User password updated, tokens invalidated"

// User tries to use old token
User App > API: Request with OLD access_token
note over API: token_version = 0 (in token)
API > JWT Middleware: Validate token
JWT Middleware > JWT Utils: Parse token
JWT Utils > JWT Middleware: Token claims (token_version: 0)
JWT Middleware > Database: Get user token_version
Database > JWT Middleware: Current token_version: 1
note over JWT Middleware: Compare: 0 != 1 
JWT Middleware > User App: 401 Unauthorized
note left of User App: "Token has been invalidated"

// User tries to refresh with old refresh token
User App > API: POST /api/v1/auth/refresh
note right of API: {refresh_token} (token_version: 0)
API > JWT Utils: Validate refresh token
JWT Utils > Database: Get user token_version
Database > JWT Utils: Current token_version: 1
note over JWT Utils: Token version mismatch (0 != 1) 
JWT Utils > User App: 401 Unauthorized
note left of User App: "Token has been invalidated. Please login again."

// User must login again
User App > API: POST /api/v1/auth/login
note right of API: {phone, password: "new_password"}
API > Database: Find user by phone
Database > API: User found (token_version: 1)
note over API: Verify new password 
API > JWT Utils: Generate new tokens (token_version: 1)
JWT Utils > API: New tokens with version 1
API > User App: 200 OK {access_token, refresh_token}
note over User App: Tokens now valid with version 1
```

---

## 6. Admin Password Update Flow (No Invalidation)

```
title Admin Password Update (Tokens Remain Valid)

// Super admin updates another admin's password
Super Admin > API: PATCH /api/v1/admin/admins/{id}
note right of API: {password: "new_password"}
API > Admin Middleware: Validate super admin token
Admin Middleware > SuperAdminOnly: Check role
SuperAdminOnly > API: Authorized (super)
API > Database: Find admin by ID
Database > API: Admin found
note over API: Hash new password (bcrypt)
note over API: NO token_version increment
API > Database: UPDATE admin SET password
Database > API: Admin updated
API > Super Admin: 200 OK "Admin password updated"

// Target admin's OLD token still works
Target Admin > API: GET /api/v1/admin/users
note right of API: Authorization: OLD admin token
API > Admin Middleware: Validate token
Admin Middleware > JWT Utils: Parse token
JWT Utils > Admin Middleware: Valid token 
note over Admin Middleware: Admin tokens don't expire
note over Admin Middleware: No token version check
Admin Middleware > API: Authorized
API > Target Admin: 200 OK (request succeeds)

note over Target Admin: Admin can continue using old token
note over Target Admin: Password changed but token still valid
```

---

## 7. User Registration Validation Flow (Flowchart)

```
// Start
User Submits Registration > Validate Phone Format
Validate Phone Format > Phone Valid?: Is E.164 format?

Phone Valid? -- Yes --> Validate Password
Phone Valid? -- No --> Return Error: "Invalid phone number format"

Validate Password > Password Long Enough?: >= 6 characters?
Password Long Enough? -- No --> Return Error: "Password must be at least 6 characters"
Password Long Enough? -- Yes --> Check Phone Exists

Check Phone Exists > Query Database: SELECT * FROM users WHERE phone = ?
Query Database > Phone Exists in DB?

Phone Exists in DB? -- Yes --> Return Error: "User with this phone already exists"
Phone Exists in DB? -- No --> Hash Password

Hash Password > bcrypt.GenerateFromPassword
bcrypt.GenerateFromPassword > Create User Record
Create User Record > Generate UUID
Generate UUID > Save to Database

Save to Database > Return Success: "User registered successfully"
Return Success > End [shape: oval]

Return Error > End [shape: oval]
```

---

## 8. Token Validation & Route Protection Flow (Flowchart)

```
// Start
Incoming Request > Protected Route?
Protected Route? -- No --> Execute Handler: Public endpoint
Execute Handler > Return Response
Return Response > End [shape: oval]

Protected Route? -- Yes --> Check Route Type
Check Route Type > User Route or Admin Route?

User Route or Admin Route? -- User --> JWT Middleware
User Route or Admin Route? -- Admin --> Admin Middleware

// User Token Validation
JWT Middleware > Extract Token: From Authorization header
Extract Token > Token Present?
Token Present? -- No --> Return 401: "Missing token"
Token Present? -- Yes --> Validate JWT Signature

Validate JWT Signature > Signature Valid?
Signature Valid? -- No --> Return 401: "Invalid token"
Signature Valid? -- Yes --> Check Token Type

Check Token Type > Is Access Token?
Is Access Token? -- No --> Return 401: "Invalid token type"
Is Access Token? -- Yes --> Check Expiration

Check Expiration > Token Expired?
Token Expired? -- Yes --> Return 401: "Token expired"
Token Expired? -- No --> Get Token Version

Get Token Version > Query User: Get user.token_version from DB
Query User > User Found?
User Found? -- No --> Return 401: "User not found"
User Found? -- Yes --> Compare Versions

Compare Versions > Token Version Matches DB?
Token Version Matches DB? -- No --> Return 401: "Token invalidated. Please login again"
Token Version Matches DB? -- Yes --> Set User Context

Set User Context > Continue to Handler: Request authorized

// Admin Token Validation
Admin Middleware > Extract Admin Token: From Authorization header
Extract Admin Token > Admin Token Present?
Admin Token Present? -- No --> Return 401: "Missing token"
Admin Token Present? -- Yes --> Validate Admin JWT

Validate Admin JWT > Admin Signature Valid?
Admin Signature Valid? -- No --> Return 401: "Invalid token"
Admin Signature Valid? -- Yes --> Check Admin Token Type

Check Admin Token Type > Is Admin Token?
Is Admin Token? -- No --> Return 401: "Invalid token type"
Is Admin Token? -- Yes --> No Expiry Check: Admin tokens are permanent

No Expiry Check > Get Admin from DB
Get Admin from DB > Admin Found?
Admin Found? -- No --> Return 401: "Admin not found"
Admin Found? -- Yes --> Check Role Requirement

Check Role Requirement > Super Admin Required?
Super Admin Required? -- Yes --> Is Super Admin?
Super Admin Required? -- No --> Set Admin Context

Is Super Admin? -- No --> Return 403: "Super admin access required"
Is Super Admin? -- Yes --> Set Admin Context

Set Admin Context > Continue to Handler: Request authorized
Continue to Handler > Execute Handler
Execute Handler > Return Response
Return 401 > End [shape: oval]
Return 403 > End [shape: oval]
```

---

## 9. System Component Relationships (Architecture)

```
title Ololo Gate - Component Dependencies

// Presentation Layer
REST API [icon: globe, color: blue] {
  Public Routes
  Protected Routes
  Admin Routes
}

// Application Layer
Handlers [icon: code, color: green] {
  User Auth Handlers
  User Management Handlers
  Admin Auth Handlers
  Admin Management Handlers
}

Middleware [icon: shield, color: yellow] {
  JWT Middleware
  Admin Middleware
  CORS Middleware
}

// Domain Layer
Models [icon: database, color: cyan] {
  User Model
  Admin Model
}

Utils [icon: tool, color: purple] {
  JWT Utilities
  Validation
}

// Infrastructure Layer
Database [icon: server, color: orange] {
  PostgreSQL
  GORM ORM
  Migrations
  Seeds
}

Config [icon: settings, color: gray] {
  Environment
  JWT Config
  DB Config
}

// Dependencies
REST API > Handlers: Routes requests
REST API > Middleware: Applies middleware
Handlers > Middleware: Uses for protection
Handlers > Models: CRUD operations
Handlers > Utils: Token generation/validation
Middleware > Utils: Token validation
Models > Database: Data persistence
Database > Config: Connection settings
Utils > Config: JWT secret & expiry
Handlers > Config: App settings
```

---

## 10. Data Flow - User Registration to First API Call

```
title Complete User Journey - Register to Authenticated Request

Flutter App [icon: mobile]
API Server [icon: server]
JWT Service [icon: key]
Database [icon: database]

// Registration
Flutter App > API Server: POST /register {phone, password}
API Server > API Server: Validate phone (E.164)
API Server > Database: Check phone exists?
Database > API Server: Not found
API Server > API Server: bcrypt.hash(password)
API Server > Database: INSERT user (token_version: 0)
Database > API Server: User created
API Server > Flutter App: 201 {id, phone}

// Login
Flutter App > API Server: POST /login {phone, password}
API Server > Database: SELECT user WHERE phone
Database > API Server: User found (token_version: 0)
API Server > API Server: bcrypt.compare(password)
API Server > JWT Service: Generate tokens (id, token_version: 0)
JWT Service > API Server: access_token, refresh_token
API Server > Flutter App: 200 {access_token, refresh_token, id}
Flutter App > Flutter App: Save tokens to secure storage

// Protected API Call
Flutter App > API Server: GET /users (Authorization: Bearer {token})
API Server > JWT Service: Validate access_token
JWT Service > JWT Service: Verify signature & expiry
JWT Service > Database: SELECT token_version FROM users WHERE id
Database > JWT Service: token_version: 0
JWT Service > JWT Service: Compare token version (0 == 0) 
JWT Service > API Server: Token valid, id
API Server > Database: SELECT * FROM users
Database > API Server: Users data
API Server > Flutter App: 200 {users: [...]}
```

---

## Notes

- **Architecture Diagram**: Shows the high-level system structure with all components
- **ERD**: Database schema with users and admins tables
- **Sequence Diagrams**: Detailed interaction flows for authentication and authorization
- **Flowcharts**: Decision trees for validation and token handling
- **Data Flow**: End-to-end journey visualization

### Key Design Decisions Visualized:

1. **Dual Authentication System**: Separate flows for users (phone-based) and admins (username-based)
2. **Token Invalidation**: User tokens have versioning, admin tokens don't
3. **Role-Based Access**: Super admins have additional privileges over regular admins
4. **Permanent Admin Tokens**: Admin tokens never expire for convenience
5. **Security Layers**: Multiple middleware checks for different authentication types

Copy each section into Eraser.io to visualize the Ololo Gate backend architecture!
