package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthData_SaveAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	authData := &AuthData{
		AccessToken:  "at",
		RefreshToken: "rt",
	}
	require.NoError(t, authData.Save())
	got, err := LoadAuthData()
	require.NoError(t, err)
	assert.Equal(t, authData.AccessToken, got.AccessToken)
	assert.Equal(t, authData.RefreshToken, got.RefreshToken)
}
