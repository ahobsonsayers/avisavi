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

	routes, err := client.RouteNetwork(
		ctx,
		cmd.Int("adults"),
		cmd.Bool("one-way"),
	)
	if err != nil {
		return err
	}

	routeList, err := routes.GetRoutes(cmd.String("origin"))
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return printJSON(routeList)
	}

	rows := make([][]string, 0, len(routeList))
	for _, route := range routeList {
		origin := fmt.Sprintf("%s (%s)", route.Origin.City, route.Origin.AirportCode)
		destination := fmt.Sprintf("%s (%s)", route.Destination.City, route.Destination.AirportCode)

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
