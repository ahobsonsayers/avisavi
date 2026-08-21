package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AuthData struct {
	AuthClientID     string `json:"x_auth_client_id"`
	MembershipNumber string `json:"membership_number"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token,omitempty"`
}

// AuthDataFilePath returns the path to the auth config file.
func AuthDataFilePath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "avios", "auth.json")
}

// Save persists the auth data to the config file.
func (a *AuthData) Save() error {
	authFilePath := AuthDataFilePath()
	err := os.MkdirAll(filepath.Dir(authFilePath), 0o700)
	if err != nil {
		return err
	}

	authJSON, err := json.Marshal(a)
	if err != nil {
		return err
	}

	return os.WriteFile(authFilePath, authJSON, 0o600)
}

// LoadAuthData reads the auth data from the config file.
func LoadAuthData() (*AuthData, error) {
	authFileBytes, err := os.ReadFile(AuthDataFilePath())
	if err != nil {
		return nil, err
	}

	var authData AuthData
	err = json.Unmarshal(authFileBytes, &authData)
	if err != nil {
		return nil, err
	}
	return &authData, nil
}
