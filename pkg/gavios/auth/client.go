package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

type Client struct {
	ClientID     string
	CallbackPort int
}

func NewClient() Client {
	return Client{
		ClientID:     os.Getenv("AVIOS_AUTH_CLIENT_ID"),
		CallbackPort: 8484,
	}
}

// Login starts the OAuth2 PKCE flow, waiting for the browser callback.
func (c Client) Login(ctx context.Context) (*AuthData, error) {
	oauthConfig := c.oauth2Config()
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
		Addr:    fmt.Sprintf(":%d", c.CallbackPort),
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

	authData := &AuthData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	_, err = authData.MembershipNumber()
	if err != nil {
		return nil, fmt.Errorf("decoding membership from token: %w", err)
	}

	return authData, nil
}

// Refresh exchanges the stored refresh token for a new access token.
func (c Client) Refresh(ctx context.Context, authData *AuthData) (*AuthData, error) {
	if authData.RefreshToken == "" {
		return nil, errors.New("no refresh token available")
	}

	oauthConfig := c.oauth2Config()
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

func (c Client) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:    c.ClientID,
		RedirectURL: fmt.Sprintf("http://localhost:%d/callback", c.CallbackPort),
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

func randHex(n int) string {
	randomBytes := make([]byte, n)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}
