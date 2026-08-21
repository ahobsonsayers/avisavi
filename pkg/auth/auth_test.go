package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeMembershipNumber(t *testing.T) {
	// JWT with claim: {"https://avios.com/membership_id":"05608372"}
	// Header: {"alg":"HS256","typ":"JWT"}
	header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	payload := "eyJodHRwczovL2F2aW9zLmNvbS9tZW1iZXJzaGlwX2lkIjoiMDU2MDgzNzIifQ"
	tok := header + "." + payload + ".fakesignature"

	mn, err := decodeMembershipNumber(tok)
	require.NoError(t, err)
	assert.Equal(t, "05608372", mn)
}

func TestDecodeMembershipNumber_Invalid(t *testing.T) {
	_, err := decodeMembershipNumber("not-a-jwt")
	assert.Error(t, err)
}
