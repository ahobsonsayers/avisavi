package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthData_MembershipNumber(t *testing.T) {
	// JWT with claim: {"https://avios.com/membership_id":"05608372"}
	// Header: {"alg":"HS256","typ":"JWT"}
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	payload := "eyJodHRwczovL2F2aW9zLmNvbS9tZW1iZXJzaGlwX2lkIjoiMDU2MDgzNzIifQ"
	token := header + "." + payload + ".ZmFrZXNpZ25hdHVyZQ"

	authData := &AuthData{AccessToken: token}
	membershipNumber, err := authData.MembershipNumber()
	require.NoError(t, err)
	assert.Equal(t, "05608372", membershipNumber)
}

func TestAuthData_MembershipNumber_Invalid(t *testing.T) {
	authData := &AuthData{AccessToken: "not-a-jwt"}
	_, err := authData.MembershipNumber()
	assert.Error(t, err)
}
