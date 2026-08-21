package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthData_SaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	td := &AuthData{
		AccessToken:      "at",
		RefreshToken:     "rt",
		MembershipNumber: "05608372",
		AuthClientID:     "BAEC-test-uuid",
	}
	require.NoError(t, td.Save())
	got, err := LoadAuthData()
	require.NoError(t, err)
	assert.Equal(t, td.AccessToken, got.AccessToken)
	assert.Equal(t, td.RefreshToken, got.RefreshToken)
	assert.Equal(t, td.MembershipNumber, got.MembershipNumber)
	assert.Equal(t, td.AuthClientID, got.AuthClientID)
}
