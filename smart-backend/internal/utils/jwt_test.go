package utils

import (
	"ololo-gate/internal/config"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupJWTTest() {
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key-for-jwt-testing",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 30 * 24 * time.Hour,
		},
	}
}

func TestGenerateTokens_Success(t *testing.T) {
	setupJWTTest()

	userID := uuid.New()
	phone := "+77771234567"
	tokenVersion := 0

	tokens, err := GenerateTokens(userID, phone, tokenVersion)

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.NotEqual(t, tokens.AccessToken, tokens.RefreshToken)
}

func TestValidateToken_AccessToken_Success(t *testing.T) {
	setupJWTTest()

	userID := uuid.New()
	phone := "+77771234567"
	tokenVersion := 0

	tokens, err := GenerateTokens(userID, phone, tokenVersion)
	assert.NoError(t, err)

	// Validate access token
	claims, err := ValidateToken(tokens.AccessToken, AccessToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, phone, claims.Phone)
	assert.Equal(t, tokenVersion, claims.TokenVersion)
	assert.Equal(t, AccessToken, claims.TokenType)
}

func TestValidateToken_RefreshToken_Success(t *testing.T) {
	setupJWTTest()

	userID := uuid.New()
	phone := "+77772345678"
	tokenVersion := 1

	tokens, err := GenerateTokens(userID, phone, tokenVersion)
	assert.NoError(t, err)

	// Validate refresh token
	claims, err := ValidateToken(tokens.RefreshToken, RefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, phone, claims.Phone)
	assert.Equal(t, tokenVersion, claims.TokenVersion)
	assert.Equal(t, RefreshToken, claims.TokenType)
}

func TestValidateToken_WrongTokenType(t *testing.T) {
	setupJWTTest()

	tokens, err := GenerateTokens(uuid.New(), "+77771234567", 0)
	assert.NoError(t, err)

	// Try to validate access token as refresh token
	_, err = ValidateToken(tokens.AccessToken, RefreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")

	// Try to validate refresh token as access token
	_, err = ValidateToken(tokens.RefreshToken, AccessToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")
}

func TestValidateToken_InvalidToken(t *testing.T) {
	setupJWTTest()

	invalidToken := "invalid.token.string"

	_, err := ValidateToken(invalidToken, AccessToken)
	assert.Error(t, err)
}

func TestValidateToken_TamperedToken(t *testing.T) {
	setupJWTTest()

	tokens, err := GenerateTokens(uuid.New(), "+77771234567", 0)
	assert.NoError(t, err)

	// Tamper with the token by adding a character
	tamperedToken := tokens.AccessToken + "x"

	_, err = ValidateToken(tamperedToken, AccessToken)
	assert.Error(t, err)
}

func TestRefreshAccessToken_Success(t *testing.T) {
	setupJWTTest()

	userID := uuid.New()
	phone := "+77771234567"
	tokenVersion := 0

	// Generate initial tokens
	tokens, err := GenerateTokens(userID, phone, tokenVersion)
	assert.NoError(t, err)

	// Use refresh token to get new access token
	newAccessToken, err := RefreshAccessToken(tokens.RefreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// Validate new access token
	claims, err := ValidateToken(newAccessToken, AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, phone, claims.Phone)
	assert.Equal(t, tokenVersion, claims.TokenVersion)
}

func TestRefreshAccessToken_InvalidRefreshToken(t *testing.T) {
	setupJWTTest()

	invalidToken := "invalid.refresh.token"

	_, err := RefreshAccessToken(invalidToken)
	assert.Error(t, err)
}

func TestRefreshAccessToken_UsingAccessToken(t *testing.T) {
	setupJWTTest()

	tokens, err := GenerateTokens(uuid.New(), "+77771234567", 0)
	assert.NoError(t, err)

	// Try to refresh using access token (should fail)
	_, err = RefreshAccessToken(tokens.AccessToken)
	assert.Error(t, err)
}

func TestTokenVersion_Included(t *testing.T) {
	setupJWTTest()

	// Test with different token versions
	testCases := []struct {
		version int
	}{
		{version: 0},
		{version: 1},
		{version: 5},
		{version: 100},
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.version)), func(t *testing.T) {
			tokens, err := GenerateTokens(uuid.New(), "+77771234567", tc.version)
			assert.NoError(t, err)

			// Validate access token contains correct version
			accessClaims, err := ValidateToken(tokens.AccessToken, AccessToken)
			assert.NoError(t, err)
			assert.Equal(t, tc.version, accessClaims.TokenVersion)

			// Validate refresh token contains correct version
			refreshClaims, err := ValidateToken(tokens.RefreshToken, RefreshToken)
			assert.NoError(t, err)
			assert.Equal(t, tc.version, refreshClaims.TokenVersion)
		})
	}
}

func TestTokenExpiry(t *testing.T) {
	// Use very short expiry for testing
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret",
			AccessExpiry:  1 * time.Nanosecond,  // Extremely short
			RefreshExpiry: 1 * time.Nanosecond,
		},
	}

	tokens, err := GenerateTokens(uuid.New(), "+77771234567", 0)
	assert.NoError(t, err)

	// Wait a bit to ensure token expires
	time.Sleep(10 * time.Millisecond)

	// Token should be expired now
	_, err = ValidateToken(tokens.AccessToken, AccessToken)
	assert.Error(t, err)
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	setupJWTTest()

	// Generate tokens for user 1
	userID1 := uuid.New()
	tokens1, err := GenerateTokens(userID1, "+77771111111", 0)
	assert.NoError(t, err)

	// Generate tokens for user 2
	userID2 := uuid.New()
	tokens2, err := GenerateTokens(userID2, "+77772222222", 0)
	assert.NoError(t, err)

	// Tokens should be different
	assert.NotEqual(t, tokens1.AccessToken, tokens2.AccessToken)
	assert.NotEqual(t, tokens1.RefreshToken, tokens2.RefreshToken)

	// Validate and check user IDs are correct
	claims1, err := ValidateToken(tokens1.AccessToken, AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID1, claims1.UserID)

	claims2, err := ValidateToken(tokens2.AccessToken, AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID2, claims2.UserID)
}
