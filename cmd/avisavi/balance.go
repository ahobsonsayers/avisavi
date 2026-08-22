package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var balanceCmd = &cli.Command{
	Name:  "balance",
	Usage: "Show your Avios balance",
	Description: `Display the current Avios balance for your account.

Examples:
  avisavi balance`,
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "json", Usage: "print raw JSON"},
	},
	Action: balanceAction,
}

func balanceAction(ctx context.Context, cmd *cli.Command) error {
	client, err := getAviosClient(ctx)
	if err != nil {
		return err
	}

	balance, err := client.Balance(ctx)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return printJSON(balance)
	}

	if balance.IsHousehold {
		fmt.Printf("%d Avios (household)\n", balance.AvailableAvios)
	} else {
		fmt.Printf("%d Avios\n", balance.AvailableAvios)
	}

	return nil
}
