package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const deepLinkPrefix = "com.usablenet.ba.avios://"

// ErrBrowserNotFound is returned when no chromium based browser is found
var ErrBrowserNotFound = errors.New("no chromium based browser found")

// findBrowser returns the path to an available chromium based browser
func findBrowser() (string, error) {
	path, ok := launcher.LookPath()
	if !ok || path == "" {
		return "", ErrBrowserNotFound
	}

	return path, nil
}

// downloadBrowser downloads chromium and returns it path
func downloadBrowser() (string, error) {
	return launcher.NewBrowser().Get()
}

// getAuthCodeViaBrowser uses a browser to get the auth code.
// It does this by opening a chroium browser on the BA login screen, waiting
// for th user to login and then capturing the auth code on redirect.
// Errors if browser cannot be found.
func getAuthCodeViaBrowser(ctx context.Context, browserPath, authURL string) (string, error) {
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
	var err error
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
