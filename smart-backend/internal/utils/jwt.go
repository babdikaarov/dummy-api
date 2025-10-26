package utils

import (
	"errors"
	"log"
	"ololo-gate/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
	AdminToken   TokenType = "admin"
)

// Claims defines the JWT claims structure
type Claims struct {
	UserID       uuid.UUID `json:"id"`
	Phone        string    `json:"phone"`
	TokenType    TokenType `json:"token_type"`
	TokenVersion int       `json:"token_version"` // Token version for invalidation
	jwt.RegisteredClaims
}

// TokenPair holds both access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GenerateTokens creates both access and refresh tokens for a user
func GenerateTokens(userID uuid.UUID, phone string, tokenVersion int) (*TokenPair, error) {
	accessExpiryMinutes := int(config.AppConfig.JWT.AccessExpiry.Minutes())
	refreshExpiryHours := int(config.AppConfig.JWT.RefreshExpiry.Hours())

	log.Printf("[TOKEN_GENERATION] Generating tokens for user ID=%s (phone=%s, token_version=%d)",
		userID, phone, tokenVersion)
	log.Printf("[TOKEN_GENERATION] Token expiry config: Access=%d minutes, Refresh=%d hours (%d days)",
		accessExpiryMinutes, refreshExpiryHours, refreshExpiryHours/24)

	// Generate access token
	accessToken, err := generateToken(userID, phone, tokenVersion, AccessToken, config.AppConfig.JWT.AccessExpiry)
	if err != nil {
		log.Printf("[TOKEN_GENERATION] Failed to generate access token: %v", err)
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := generateToken(userID, phone, tokenVersion, RefreshToken, config.AppConfig.JWT.RefreshExpiry)
	if err != nil {
		log.Printf("[TOKEN_GENERATION] Failed to generate refresh token: %v", err)
		return nil, err
	}

	log.Printf("[TOKEN_GENERATION] ✅ Tokens generated successfully. Access token expires in %d minutes, Refresh token expires in %d hours (%d days)",
		accessExpiryMinutes, refreshExpiryHours, refreshExpiryHours/24)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateToken creates a JWT token with the specified parameters
func generateToken(userID uuid.UUID, phone string, tokenVersion int, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	// Calculate expiry in minutes for logging
	expiryMinutes := int(expiry.Minutes())
	expiryHours := int(expiry.Hours())
	expiryDays := expiryHours / 24

	claims := Claims{
		UserID:       userID,
		Phone:        phone,
		TokenType:    tokenType,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		log.Printf("[TOKEN_GENERATION] Failed to sign %s token: %v", tokenType, err)
		return "", err
	}

	// Log token details
	if expiryDays > 0 {
		log.Printf("[TOKEN_INFO] %s token created: User=%s, Phone=%s, token_version=%d, IssuedAt=%s, ExpiresAt=%s (in %d days, %d hours)",
			tokenType, userID, phone, tokenVersion, now.Format("2006-01-02 15:04:05"), expiresAt.Format("2006-01-02 15:04:05"), expiryDays, expiryHours%24)
	} else if expiryHours > 0 {
		log.Printf("[TOKEN_INFO] %s token created: User=%s, Phone=%s, token_version=%d, IssuedAt=%s, ExpiresAt=%s (in %d hours, %d minutes)",
			tokenType, userID, phone, tokenVersion, now.Format("2006-01-02 15:04:05"), expiresAt.Format("2006-01-02 15:04:05"), expiryHours, expiryMinutes%60)
	} else {
		log.Printf("[TOKEN_INFO] %s token created: User=%s, Phone=%s, token_version=%d, IssuedAt=%s, ExpiresAt=%s (in %d minutes)",
			tokenType, userID, phone, tokenVersion, now.Format("2006-01-02 15:04:05"), expiresAt.Format("2006-01-02 15:04:05"), expiryMinutes)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		log.Printf("[TOKEN_VALIDATION] Token validation failed: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Printf("[TOKEN_VALIDATION] Token claims invalid or token not valid")
		return nil, errors.New("invalid token")
	}

	// Verify token type
	if claims.TokenType != expectedType {
		log.Printf("[TOKEN_VALIDATION] Token type mismatch. Expected=%s, Got=%s", expectedType, claims.TokenType)
		return nil, errors.New("invalid token type")
	}

	// Log token info
	now := time.Now()
	expiresAt := claims.ExpiresAt.Time
	timeUntilExpiry := expiresAt.Sub(now)
	minutesUntilExpiry := int(timeUntilExpiry.Minutes())
	hoursUntilExpiry := int(timeUntilExpiry.Hours())
	daysUntilExpiry := hoursUntilExpiry / 24

	if daysUntilExpiry > 0 {
		log.Printf("[TOKEN_INFO] %s token validated: User ID=%s, Phone=%s, token_version=%d, ExpiresAt=%s (in %d days, %d hours)",
			claims.TokenType, claims.UserID, claims.Phone, claims.TokenVersion, expiresAt.Format("2006-01-02 15:04:05"), daysUntilExpiry, hoursUntilExpiry%24)
	} else if hoursUntilExpiry > 0 {
		log.Printf("[TOKEN_INFO] %s token validated: User ID=%s, Phone=%s, token_version=%d, ExpiresAt=%s (in %d hours, %d minutes)",
			claims.TokenType, claims.UserID, claims.Phone, claims.TokenVersion, expiresAt.Format("2006-01-02 15:04:05"), hoursUntilExpiry, minutesUntilExpiry%60)
	} else {
		log.Printf("[TOKEN_INFO] %s token validated: User ID=%s, Phone=%s, token_version=%d, ExpiresAt=%s (in %d minutes)",
			claims.TokenType, claims.UserID, claims.Phone, claims.TokenVersion, expiresAt.Format("2006-01-02 15:04:05"), minutesUntilExpiry)
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(refreshTokenString string) (string, error) {
	log.Printf("[TOKEN_REFRESH] Starting token refresh process...")

	// Validate refresh token
	claims, err := ValidateToken(refreshTokenString, RefreshToken)
	if err != nil {
		log.Printf("[TOKEN_REFRESH] Refresh token validation failed: %v", err)
		return "", err
	}

	log.Printf("[TOKEN_REFRESH] Refresh token validated. User ID=%s, Phone=%s, token_version=%d",
		claims.UserID, claims.Phone, claims.TokenVersion)

	// Generate new access token with the same token version
	accessToken, err := generateToken(claims.UserID, claims.Phone, claims.TokenVersion, AccessToken, config.AppConfig.JWT.AccessExpiry)
	if err != nil {
		log.Printf("[TOKEN_REFRESH] Failed to generate new access token: %v", err)
		return "", err
	}

	accessExpiryMinutes := int(config.AppConfig.JWT.AccessExpiry.Minutes())
	log.Printf("[TOKEN_REFRESH] ✅ New access token generated successfully. Expires in %d minutes",
		accessExpiryMinutes)

	return accessToken, nil
}

// AdminClaims defines the JWT claims structure for admin tokens
type AdminClaims struct {
	AdminID      uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`        // "super" or "regular"
	TokenType    TokenType `json:"token_type"`   // always "admin"
	TokenVersion int       `json:"token_version"` // Token version for invalidation
	jwt.RegisteredClaims
}

// GenerateAdminToken creates a permanent JWT token for admins (no expiry)
func GenerateAdminToken(adminID uuid.UUID, username, role string, tokenVersion int) (string, error) {
	log.Printf("[TOKEN_GENERATION] Generating admin token for Admin ID=%s (username=%s, role=%s, token_version=%d)",
		adminID, username, role, tokenVersion)

	now := time.Now()
	claims := AdminClaims{
		AdminID:      adminID,
		Username:     username,
		Role:         role,
		TokenType:    AdminToken,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			// No ExpiresAt - token never expires
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.Secret))
	if err != nil {
		log.Printf("[TOKEN_GENERATION] Failed to sign admin token: %v", err)
		return "", err
	}

	log.Printf("[TOKEN_INFO] Admin token created: Admin ID=%s, Username=%s, Role=%s, token_version=%d, IssuedAt=%s (NEVER EXPIRES)",
		adminID, username, role, tokenVersion, now.Format("2006-01-02 15:04:05"))

	return tokenString, nil
}

// ValidateAdminToken validates an admin JWT token and returns the claims
func ValidateAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(config.AppConfig.JWT.Secret), nil
	})

	if err != nil {
		log.Printf("[TOKEN_VALIDATION] Admin token validation failed: %v", err)
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		log.Printf("[TOKEN_VALIDATION] Admin token claims invalid or token not valid")
		return nil, errors.New("invalid token")
	}

	// Verify token type
	if claims.TokenType != AdminToken {
		log.Printf("[TOKEN_VALIDATION] Admin token type mismatch. Expected=%s, Got=%s", AdminToken, claims.TokenType)
		return nil, errors.New("invalid token type")
	}

	// Log admin token info
	issuedAt := claims.IssuedAt.Time
	log.Printf("[TOKEN_INFO] Admin token validated: Admin ID=%s, Username=%s, Role=%s, token_version=%d, IssuedAt=%s (NEVER EXPIRES)",
		claims.AdminID, claims.Username, claims.Role, claims.TokenVersion, issuedAt.Format("2006-01-02 15:04:05"))

	return claims, nil
}
