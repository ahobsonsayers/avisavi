package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/ahobsonsayers/avios-cli/pkg/gavios/auth"
	"github.com/urfave/cli/v3"
)

var loginCmd = &cli.Command{
	Name:  "login",
	Usage: "Log in to Avios via browser",
	Description: `Authenticate with your Avios account.

Opens a browser to complete auth flow, then saves the token
to ~/.config/avios/auth.json. The membership number is decoded from the
JWT automatically.

Examples:
  avios login
  avios login --client-id LO0m9CsTZZ9qY9zY9DD2JdngeR76qqND`,

	Flags: []cli.Flag{
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

	authData, err := client.Login(ctx)
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
