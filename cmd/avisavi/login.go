package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/ahobsonsayers/avisavi/pkg/gavios/auth"
	cloak "github.com/enowdev/cloak-go"
	"github.com/urfave/cli/v3"
)

var loginCmd = &cli.Command{
	Name:  "login",
	Usage: "Log in to Avios via browser",
	Description: `Authenticate with your Avios account.

Opens a browser to complete the auth flow, then saves the token.
The membership number is decoded from the JWT automatically.

For this to work, CloakBrowser must be used due to its ability
to sidestep anti-automation protection. It will be downloaded
on first run (~200MB). Use 'avisavi login cleanup' to remove it.

If you would prefer not to download or use a browser, you can
use manual mode to paste the redirect URL yourself:
  avisavi login --manual

Examples:
  avisavi login
  avisavi login --manual
  avisavi login --client-id XXXXX
  avisavi login cleanup`,

	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "manual", Usage: "skip browser automation, paste the redirect URL manually"},
		&cli.StringFlag{Name: "client-id", Usage: "Auth0 client ID (overrides AVIOS_AUTH_CLIENT_ID env)"},
	},

	Commands: []*cli.Command{
		{
			Name:  "cleanup",
			Usage: "Delete the downloaded CloakBrowser binary",
			Description: `Removes the downloaded CloakBrowser binary (~200MB), used for
logging in. It will be re-downloaded on the next login.

Examples:
  avisavi login cleanup`,
			Action: loginCleanupAction,
		},
	},

	Action: loginAction,
}

func loginAction(ctx context.Context, cmd *cli.Command) error {
	client := auth.NewClient()

	clientID := cmd.String("client-id")
	if clientID != "" {
		client.ClientID = clientID
	}

	if client.ClientID == "" {
		return errors.New("AVIOS_AUTH_CLIENT_ID must be set\n  use --client-id or export AVIOS_AUTH_CLIENT_ID")
	}

	var mode auth.AuthMode
	if cmd.Bool("manual") {
		mode = auth.Manual
	} else {
		fmt.Println("Logging in using CloakBrowser.")
		fmt.Println("If not found, it will be downloaded first (~200MB - be patient)")
		mode = auth.CloakBrowser
	}

	authData, err := client.Login(ctx, mode)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	err = authData.Save()
	if err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	fmt.Println("Logged in. Token saved to", auth.AuthDataFilePath())
	return nil
}

func loginCleanupAction(ctx context.Context, cmd *cli.Command) error {
	browserDir := cloak.CacheDir()

	_, err := os.Stat(browserDir)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Nothing to clean up. CloakBrowser not found at %s\n", browserDir)
		return nil
	}

	err = os.RemoveAll(browserDir)
	if err != nil {
		return fmt.Errorf("failed to remove CloakBrowser: %w", err)
	}

	fmt.Printf("Removed CloakBrowser at %s\n", browserDir)
	return nil
}
