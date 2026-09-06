package main

import (
	"context"
	"fmt"
	"strconv"

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
  avisavi routes --destination JFK --destination BOS
  avisavi routes --region "North America"`,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "origin",
			Aliases: []string{"o"},
			Usage:   "IATA origin code (repeatable, scans all origins if omitted)",
		},
		&cli.StringSliceFlag{
			Name:    "destination",
			Aliases: []string{"d"},
			Usage:   "IATA destination code (repeatable, scans all destinations if omitted)",
		},
		&cli.StringSliceFlag{
			Name:  "region",
			Usage: "destination region e.g. \"North America\" (repeatable)",
		},
		&cli.BoolFlag{Name: "json", Usage: "print raw JSON"},
	},
	Action: routesAction,
}

func routesAction(ctx context.Context, cmd *cli.Command) error {
	client, err := getAviosClient(ctx)
	if err != nil {
		return err
	}

	network, err := client.Network(ctx)
	if err != nil {
		return err
	}

	routes, err := network.FindRoutes(gavios.FindRoutesInput{
		Origins:            cmd.StringSlice("origin"),
		Destinations:       cmd.StringSlice("destination"),
		DestinationRegions: cmd.StringSlice("region"),
	})
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return printJSON(routes)
	}

	rows := make([][]string, 0, len(routes))
	for _, route := range routes {
		origin := fmt.Sprintf("%s (%s)", route.Origin.City, originCodeStyle.Render(route.Origin.AirportCode))
		destination := fmt.Sprintf(
			"%s (%s)", route.Destination.City, destinationCodeStyle.Render(route.Destination.AirportCode),
		)

		economyPrice := aviosRange(route.Details.AviosPrices.Economy)
		businessPrice := aviosRange(route.Details.AviosPrices.Business)

		rows = append(
			rows,
			[]string{origin, "→", destination, economyPrice, businessPrice},
		)
	}

	table := baseTable.
		Headers("Origin", "", "Destination", "Economy", "Business").
		Rows(rows...)

	fmt.Println(table.Render())
	return nil
}

// aviosRange formats an Avios price range with thousands separators.
func aviosRange(prices gavios.AviosPrice) string {
	return fmt.Sprintf("%s - %s", groupedNumber(prices.MinAvios), groupedNumber(prices.MaxAvios))
}

// groupedNumber formats a number with thousands separators.
func groupedNumber(value int) string {
	digits := strconv.Itoa(value)
	grouped := ""
	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped += ","
		}
		grouped += string(digit)
	}
	return grouped
}
