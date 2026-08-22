package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ahobsonsayers/avios-cli/pkg/gavios"
	"github.com/ahobsonsayers/avios-cli/pkg/gavios/auth"
	"github.com/urfave/cli/v3"
)

var rootCmd = &cli.Command{
	Name:  "avios",
	Usage: "Avios Reward Flight CLI",
	Description: `Search and check Avios reward flights from the terminal.

Before using any command, run 'avios login' to authenticate.
Configuration is read from environment variables.

Examples:
  avios login
  avios balance
  avios routes --origin LON
  avios availability --origin LON --destination NYC
  avios search --origin LON --outbound 2026-09-09 --return 2026-09-13`,
	Commands: []*cli.Command{
		loginCmd,
		balanceCmd,
		routesCmd,
		availabilityCmd,
		searchCmd,
	},
}

func main() {
	err := rootCmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getAviosClient(ctx context.Context) (*gavios.Client, error) {
	authData, err := auth.LoadAuthData()
	if err != nil {
		return nil, fmt.Errorf("not logged in — run 'avios login' first: %w", err)
	}

	if authData.NeedsRefresh() {
		authClient := auth.NewClient()
		authData, err = authClient.Refresh(ctx, authData)
		if err != nil {
			return nil, fmt.Errorf("session expired — run 'avios login' again: %w", err)
		}

		err = authData.Save()
		if err != nil {
			return nil, fmt.Errorf("saving refreshed token: %w", err)
		}
	}

	return gavios.NewClient(authData), nil
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
