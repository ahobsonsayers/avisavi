package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"text/tabwriter"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
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

	originFlag := cmd.String("origin")
	if originFlag != "" {
		originCode, err := gavios.NormalizeAirportCode(originFlag)
		if err != nil {
			return err
		}

		originRoutes, found := routes.Routes[originCode]
		if !found {
			return fmt.Errorf("origin %q not found in routes", originCode)
		}

		airports := map[string]gavios.Airport{originCode: routes.Airports[originCode]}
		for destinationCode := range originRoutes {
			airports[destinationCode] = routes.Airports[destinationCode]
		}

		routes = gavios.Routes{
			Airports: airports,
			Routes:   map[string]map[string]gavios.Route{originCode: originRoutes},
		}
	}

	if cmd.Bool("json") {
		return printJSON(routes)
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, originCode := range slices.Sorted(maps.Keys(routes.Routes)) {
		origin := routes.Airports[originCode]
		for _, destinationCode := range slices.Sorted(maps.Keys(routes.Routes[originCode])) {
			destination := routes.Airports[destinationCode]
			route := routes.Routes[originCode][destinationCode]
			fmt.Fprintf(
				writer,
				"%s (%s)\t->\t%s (%s)\tEconomy %d-%d\tBusiness %d-%d\n",
				origin.City, origin.Code,
				destination.City, destination.Code,
				route.AviosPrices.Economy.MinAvios, route.AviosPrices.Economy.MaxAvios,
				route.AviosPrices.Business.MinAvios, route.AviosPrices.Business.MaxAvios,
			)
		}
	}

	return writer.Flush()
}
