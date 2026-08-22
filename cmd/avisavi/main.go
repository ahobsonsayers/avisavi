package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
	"github.com/ahobsonsayers/avisavi/pkg/gavios/auth"
	"github.com/urfave/cli/v3"
)

var rootCmd = &cli.Command{
	Name:  "avisavi",
	Usage: "Avios Reward Flight CLI",
	Description: `Search and check Avios reward flights from the terminal.

Before using any command, run 'avisavi login' to authenticate.
Configuration is read from environment variables.

Examples:
  avisavi login
  avisavi balance
  avisavi routes --origin LON
  avisavi availability --origin LON --destination NYC
  avisavi search --origin LON --outbound 2026-09-09 --return 2026-09-13`,
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
		return nil, fmt.Errorf("not logged in — run 'avisavi login' first: %w", err)
	}

	if authData.NeedsRefresh() {
		authClient := auth.NewClient()
		authData, err = authClient.Refresh(ctx, authData)
		if err != nil {
			return nil, fmt.Errorf("session expired — run 'avisavi login' again: %w", err)
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
