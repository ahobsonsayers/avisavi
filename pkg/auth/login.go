package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// Login starts the OAuth2 PKCE flow, waiting for the browser callback.
func Login(ctx context.Context, cfg Config) (*AuthData, error) {
	oauthConfig := oauth2Config(cfg)
	verifier := oauth2.GenerateVerifier()
	state := randHex(16)

	authURL := oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("audience", "https://api.avios.com/"),
		oauth2.S256ChallengeOption(verifier),
	)

	// Local callback server captures the auth code and state from the redirect.
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	authHandlerFunc := func(response http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		if query.Get("state") != state {
			http.Error(response, "state mismatch", http.StatusBadRequest)
			errChan <- errors.New("oauth state mismatch")
			return
		}

		codeChan <- query.Get("code")
		fmt.Fprintln(response, "<h1>Done! You can close this tab.</h1>")
	}

	authServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: http.HandlerFunc(authHandlerFunc),
	}

	// Start auth redirect server
	go func() { _ = authServer.ListenAndServe() }()
	defer func() { _ = authServer.Shutdown(ctx) }()

	log.Println("Opening browser for login...")
	log.Printf("If it doesn't open, visit:\n%s\n", authURL)
	err := browser.OpenURL(authURL)
	if err != nil {
		log.Printf("could not open browser: %v — visit the URL above manually\n", err)
	}

	// Exchange the auth code for tokens, or fail on mismatch/timeout.
	var token *oauth2.Token
	select {
	case code := <-codeChan:
		var err error
		token, err = oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			return nil, fmt.Errorf("token exchange: %w", err)
		}

	case err := <-errChan:
		return nil, err

	case <-time.After(5 * time.Minute):
		return nil, errors.New("login timed out waiting for callback")
	}

	membershipNumber, err := decodeMembershipNumber(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("decoding membership from token: %w", err)
	}

	authData := &AuthData{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		MembershipNumber: membershipNumber,
		AuthClientID:     generateAuthClientID(),
	}

	return authData, nil
}

// Refresh exchanges the stored refresh token for a new access token.
func Refresh(ctx context.Context, cfg Config, authData *AuthData) (*AuthData, error) {
	if authData.RefreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	oauthConfig := oauth2Config(cfg)
	tokenSource := oauthConfig.TokenSource(
		ctx,
		&oauth2.Token{
			RefreshToken: authData.RefreshToken,
		},
	)
	token, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	authData.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		authData.RefreshToken = token.RefreshToken
	}

	return authData, nil
}

func oauth2Config(cfg Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:    cfg.ClientID,
		RedirectURL: cfg.Callback(),
		Scopes: []string{
			"openid", "profile", "email",
			"read:transaction", "read:member", "read:account",
			"offline_access",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.britishairways.com/authorize",
			TokenURL: "https://accounts.britishairways.com/oauth/token",
		},
	}
}

// generateAuthClientID builds a per-login client ID for the x-auth-client-id header.
func generateAuthClientID() string {
	return "BAEC-" + uuid.NewString()
}

func randHex(n int) string {
	randomBytes := make([]byte, n)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}
