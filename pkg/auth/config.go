package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ClientID string
	Port     int
}

func NewConfig() Config {
	return Config{
		ClientID: os.Getenv("AVIOS_AUTH_CLIENT_ID"),
		Port:     8484,
	}
}

func (c Config) Callback() string {
	return fmt.Sprintf("http://localhost:%d/callback", c.Port)
}

func decodeMembershipNumber(accessToken string) (string, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) < 2 {
		return "", errors.New("invalid JWT: not enough segments")
	}
	payload := parts[1]
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("decoding JWT payload: %w", err)
	}
	var claims struct {
		MembershipID string `json:"https://avios.com/membership_id"`
	}
	err = json.Unmarshal(decoded, &claims)
	if err != nil {
		return "", fmt.Errorf("parsing JWT claims: %w", err)
	}
	if claims.MembershipID == "" {
		return "", errors.New("membership_id not found in JWT claims")
	}
	return claims.MembershipID, nil
}
