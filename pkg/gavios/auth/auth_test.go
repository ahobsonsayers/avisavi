package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthData_MembershipNumber(t *testing.T) {
	// JWT with claim: {"https://avios.com/membership_id":"01234567"}
	// Header: {"alg":"HS256","typ":"JWT"}
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	payload := "eyJodHRwczovL2F2aW9zLmNvbS9tZW1iZXJzaGlwX2lkIjoiMDEyMzQ1NjcifQ"
	token := header + "." + payload + ".ZmFrZXNpZ25hdHVyZQ"

	authData := &AuthData{AccessToken: token}
	membershipNumber, err := authData.MembershipNumber()
	require.NoError(t, err)
	assert.Equal(t, "01234567", membershipNumber)
}

func TestAuthData_MembershipNumber_Invalid(t *testing.T) {
	authData := &AuthData{AccessToken: "not-a-jwt"}
	_, err := authData.MembershipNumber()
	assert.Error(t, err)
}

func TestAuthData_NeedsRefresh(t *testing.T) {
	makeToken := func(claims map[string]any) string {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
		payload := base64.RawURLEncoding.EncodeToString(mustJSON(t, claims))
		return header + "." + payload + ".ZmFrZXNpZ25hdHVyZQ"
	}

	tests := []struct {
		name   string
		token  string
		expect bool
	}{
		{"empty token", "", true},
		{"not a jwt", "not-a-jwt", true},
		{"no exp claim", makeToken(map[string]any{}), true},
		{"expired", makeToken(map[string]any{"exp": time.Now().Add(-time.Hour).Unix()}), true},
		{"unexpired", makeToken(map[string]any{"exp": time.Now().Add(time.Hour).Unix()}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authData := &AuthData{AccessToken: tt.token}
			assert.Equal(t, tt.expect, authData.NeedsRefresh())
		})
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return encoded
}
