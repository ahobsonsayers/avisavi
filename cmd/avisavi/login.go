package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahobsonsayers/avisavi/pkg/gavios/auth"
	"github.com/urfave/cli/v3"
)

var loginCmd = &cli.Command{
	Name:  "login",
	Usage: "Log in to Avios via browser",
	Description: `Authenticate with your Avios account.

Opens a browser to complete auth flow, then saves the token.
The membership number is decoded from the JWT automatically.

For this to work, CloakBrowser must be used due to its ability
to sidestep anti-automation protection. It will be downloaded
on first run (~200MB).

If you would prefer not to download or use a browser, you can
use manual mode to paste the redirect URL yourself:
  avisavi login --manual

Examples:
  avisavi login
  avisavi login --manual
  avisavi login --client-id XXXXX`,

	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "manual", Usage: "skip browser automation, paste the redirect URL manually"},
		&cli.StringFlag{Name: "client-id", Usage: "Auth0 client ID (overrides AVIOS_AUTH_CLIENT_ID env)"},
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
		fmt.Println("If not found, it will be downloaded first (~200Mb - be patient).")
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
