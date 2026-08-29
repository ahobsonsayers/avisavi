package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// defaultClientID is embedded at build time via:
// -ldflags "-X github.com/ahobsonsayers/avisavi/pkg/gavios/auth.defaultClientID=<id>".
// It is a fallback for binaries that lack the AVIOS_AUTH_CLIENT_ID env var at runtime.
var defaultClientID string

const (
	auth0Domain = "accounts.britishairways.com"
	audience    = "https://api.avios.com/"
	redirectURI = "com.usablenet.ba.avios://accounts.britishairways.com/android/com.usablenet.ba.avios/callback"
)

type Client struct {
	ClientID string
}

func NewClient() Client {
	clientID := os.Getenv("AVIOS_AUTH_CLIENT_ID")
	if clientID == "" {
		clientID = defaultClientID
	}

	return Client{
		ClientID: clientID,
	}
}

type AuthMode int

const (
	// Browser uses an installed Chromium browser to login.
	// If Chromium is not found, login fails.
	Browser AuthMode = iota

	// BrowserWithDownload uses an installed Chromium browser to login,
	// and if Chromium is not found, it is downloaded and used
	BrowserWithDownload

	// Manual logs in using manual paste of redirect uri
	Manual
)

func (c Client) Login(ctx context.Context, authMode AuthMode) (*AuthData, error) {
	oauthConfig := c.oauth2Config()
	verifier := oauth2.GenerateVerifier()

	authURL := oauthConfig.AuthCodeURL(
		"",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("audience", audience),
		oauth2.S256ChallengeOption(verifier),
	)

	var authcode string
	var err error
	if authMode == Manual {
		authcode, err = getAuthCodeManually(authURL)
	} else {
		// Get browser path
		var browserPath string
		if authMode == BrowserWithDownload {
			browserPath, err = downloadBrowser()
			if err != nil {
				return nil, err
			}
		} else { // Browser
			browserPath, err = findBrowser()
			if err != nil {
				return nil, err
			}
		}
		authcode, err = getAuthCodeViaBrowser(ctx, browserPath, authURL)
	}
	if err != nil {
		return nil, err
	}
	if authcode == "" {
		return nil, errors.New("no auth code captured")
	}

	token, err := oauthConfig.Exchange(ctx, authcode, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
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
		RedirectURL: redirectURI,
		Scopes: []string{
			"openid", "profile", "email",
			"read:transaction", "read:member", "read:account",
			"offline_access",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://" + auth0Domain + "/authorize",
			TokenURL: "https://" + auth0Domain + "/oauth/token",
		},
	}
}

// getAuthCodeManually opens the system browser to the authorize URL and
// prompts the user to paste the deep link redirect uri after logging in.
func getAuthCodeManually(authURL string) (string, error) {
	err := browser.OpenURL(authURL)
	if err != nil {
		return "", fmt.Errorf("failed to open browser: %w", err)
	}

	fmt.Println("After logging in, the browser will attempt to redirect and fail.")
	fmt.Println("This may look like the browser does nothing after clicking login.")
	fmt.Println()
	fmt.Println("If the browser redirects, copy the URL from the address bar.")
	fmt.Println("If it does not redirect, open DevTools with F12, go to the Network tab,")
	fmt.Println("click login again, and copy the URL from the 'Location' header of the")
	fmt.Println("request that fires.")
	fmt.Println()
	fmt.Println("In either case the URL will start with com.usablenet.ba.avios://...")
	fmt.Println()
	fmt.Print("Paste the URL here: ")

	inputReader := bufio.NewReader(os.Stdin)
	inputString, err := inputReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	inputString = strings.TrimSpace(inputString)
	if inputString == "" {
		return "", fmt.Errorf("input is empty")
	}

	parsedUrl, err := url.Parse(inputString)
	if err != nil {
		return "", fmt.Errorf("failed to parse input as a url: %w", err)
	}

	authCode := parsedUrl.Query().Get("code")
	return authCode, nil
}
