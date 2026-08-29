package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v3"
)

var routesCmd = &cli.Command{
	Name:  "routes",
	Usage: "List reward-flight routes with Avios price ranges",
	Description: `Fetch reward-flight routes, showing
the minimum and maximum Avios needed per cabin class.

Examples:
  avisavi routes
  avisavi routes --origin LON
  avisavi routes --origin JFK --adults 2 --one-way`,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "origin", Aliases: []string{"o"}, Usage: "IATA origin code (omit for all origins)"},
		&cli.IntFlag{Name: "adults", Aliases: []string{"a"}, Value: 1, Usage: "number of adults"},
		&cli.BoolFlag{Name: "one-way", Usage: "one-way flights only"},
		&cli.BoolFlag{Name: "json", Usage: "print raw JSON"},
	},
	Action: routesAction,
}

func routesAction(ctx context.Context, cmd *cli.Command) error {
	client, err := getAviosClient(ctx)
	if err != nil {
		return err
	}

	routes, err := client.Routes(
		ctx,
		cmd.Int("adults"),
		cmd.Bool("one-way"),
	)
	if err != nil {
		return err
	}

	routes, err = routes.RoutesFromOrigin(cmd.String("origin"))
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return printJSON(routes)
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, originAirport := range routes.Origins {
		for _, destination := range originAirport.Destinations {
			economy := destination.AviosPerCabinClass["Economy"]
			business := destination.AviosPerCabinClass["Business"]
			fmt.Fprintf(
				writer,
				"%s -> %s\t%s, %s\tEco %d-%d\tBus %d-%d\n",
				originAirport.AirportCode, destination.DestinationAirportCode,
				destination.DestinationName, destination.CountryName,
				economy.Min, economy.Max, business.Min, business.Max)
		}
	}

	return writer.Flush()
}
