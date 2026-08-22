package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-jwt/jwt/v5"
)

type AuthData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// MembershipNumber decodes the membership number from the access token.
func (a *AuthData) MembershipNumber() (string, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(a.AccessToken, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("parsing token claims")
	}

	membershipNumber, ok := claims["https://avios.com/membership_id"].(string)
	if !ok || membershipNumber == "" {
		return "", errors.New("membership_id not found in token claims")
	}

	return membershipNumber, nil
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
