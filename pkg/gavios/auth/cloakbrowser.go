package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	cloak "github.com/enowdev/cloak-go"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const deepLinkPrefix = "com.usablenet.ba.avios://"

// getAuthCodeViaCloakBrowser uses cloakbrowser to get the auth code.
// It does this by opening cloakbrowser on the BA login screen, waiting
// for the user to login and then capturing the auth code on redirect.
// cloakbrowser is required for stealth capability, if it is missing it is downloaded.
func getAuthCodeViaCloakBrowser(ctx context.Context, authURL string) (string, error) {
	// Get cloakbrowser path, downloading it if it is not found
	browserPath, err := cloakBrowserPath(ctx)
	if err != nil {
		return "", err
	}

	// Run browser in fresh profile
	controlURL := launcher.New().
		Bin(browserPath).
		Headless(false).
		Set("no-first-run").
		Set("no-default-browser-check").
		MustLaunch()

	// Open browser and pages
	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("")

	// Context with timeout, and that can be used to signal login is done
	loginCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Start go routine that waits for the redirect, so
	// it can capture the auth code in the url
	var authCode string
	go func() {
		wait := page.EachEvent(
			func(event *proto.NetworkRequestWillBeSent) bool {
				requestURL := event.Request.URL

				// Check if the url in request is the deep link redirect uri
				// If not wait for next network request event
				if !strings.HasPrefix(requestURL, deepLinkPrefix) {
					return false
				}

				parsedUrl, parseErr := url.Parse(requestURL)
				if parseErr != nil {
					err = fmt.Errorf("failed to parse url: %w", parseErr)
					return true
				}

				// Parse out the auth code
				authCode = parsedUrl.Query().Get("code")
				return true
			},
		)
		wait()
		cancel()
	}()

	// Open login page
	page.MustNavigate(authURL)

	// Wait for login ctx to be done
	<-loginCtx.Done()
	switch loginCtx.Err() {
	case context.Canceled:
		return authCode, err
	case context.DeadlineExceeded:
		return "", errors.New("login timed out")
	default:
		return "", loginCtx.Err()
	}
}

// cloakBrowserPath gets the path of cloakbrowser, downloading it if it is missing
func cloakBrowserPath(ctx context.Context) (string, error) {
	browserPath, err := cloak.EnsureBinary(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get cloakbrowser path: %w", err)
	}

	return browserPath, nil
}
