package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

var searchCmd = &cli.Command{
	Name:  "search",
	Usage: "Find destinations with seats on specific outbound and return dates",
	Description: `Search all reward destinations from an origin for available seats
on the given outbound and return dates.

This fetches the route list, then checks availability for each destination.

Examples:
  avisavi search --origin LON --outbound 2026-09-09 --return 2026-09-13
  avisavi search --origin LON --outbound 2026-09-09 --return 2026-09-13 --cabin Business
  avisavi search --origin LON --outbound 2026-09-09 --return 2026-09-13 --adults 2 --min-seats 4`,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "origin", Aliases: []string{"o"}, Value: "LON", Usage: "IATA origin code"},
		&cli.StringFlag{Name: "outbound", Required: true, Usage: "outbound date YYYY-MM-DD"},
		&cli.StringFlag{Name: "return", Required: true, Usage: "return date YYYY-MM-DD"},
		&cli.IntFlag{Name: "adults", Aliases: []string{"a"}, Value: 1, Usage: "number of adults"},
		&cli.StringFlag{Name: "cabin", Usage: "filter by cabin (Economy, Premium, Business, First)"},
		&cli.IntFlag{Name: "min-seats", Value: 1, Usage: "minimum seats required on each leg"},
		&cli.BoolFlag{Name: "json", Usage: "print raw JSON"},
	},
	Action: searchAction,
}

type searchResult struct {
	destination  gavios.Airport
	availability gavios.RouteFlights
}

func searchAction(ctx context.Context, cmd *cli.Command) error {
	client, err := getAviosClient(ctx)
	if err != nil {
		return err
	}

	origin := cmd.String("origin")
	adults := cmd.Int("adults")
	outbound := cmd.String("outbound")
	returnDate := cmd.String("return")
	cabin := cmd.String("cabin")
	minSeats := cmd.Int("min-seats")

	results := searchDestinations(ctx, client, origin, adults)
	if cmd.Bool("json") {
		allRouteFlights := make([]gavios.RouteFlights, 0, len(results))
		for _, result := range results {
			allRouteFlights = append(allRouteFlights, result.availability)
		}
		return printJSON(allRouteFlights)
	}

	return renderSearchResults(os.Stdout, results, outbound, returnDate, cabin, minSeats)
}

func searchDestinations(ctx context.Context, client *gavios.Client, originCode string, adults int) []searchResult {
	routes, err := client.RouteNetwork(ctx, adults, false)
	if err != nil {
		return nil
	}

	routeList, err := routes.GetRoutes(originCode)
	if err != nil {
		return nil
	}

	results := make([]searchResult, 0, len(routeList))
	for _, route := range routeList {
		routeFlights, err := client.RouteFlights(
			ctx,
			route.Origin.AirportCode,
			route.Destination.AirportCode,
			false,
			adults,
		)
		if err != nil {
			continue
		}

		results = append(
			results,
			searchResult{
				destination:  route.Destination,
				availability: routeFlights,
			},
		)
	}
	return results
}

func renderSearchResults(w io.Writer, results []searchResult, outbound, returnDate, cabin string, minSeats int) error {
	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprint(writer, "DEST\tNAME\tCOUNTRY\tCABIN\tOUT\tSEATS\tRET\tSEATS\n")

	// The API keys flights by date but each flight's own date field may
	// differ from the key, so match on the requested date prefix.

	for _, result := range results {
		cabins := []struct {
			name    string
			flights gavios.TripFlights
		}{
			{"Economy", result.availability.Economy},
			{"Premium", result.availability.Premium},
			{"Business", result.availability.Business},
			{"First", result.availability.First},
		}

		for _, cabinTripFlights := range cabins {
			if cabin != "" && cabinTripFlights.name != cabin {
				continue
			}

			cabinName := cabinTripFlights.name
			tripFlights := cabinTripFlights.flights
			outboundFlights := filterSeats(flightsOnDate(tripFlights.Outbound, outbound), minSeats)
			inboundFlights := filterSeats(flightsOnDate(tripFlights.Inbound, returnDate), minSeats)
			if len(outboundFlights) > 0 && len(inboundFlights) > 0 {
				fmt.Fprintf(
					writer, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\n",
					result.destination.AirportCode,
					result.destination.City,
					result.destination.Country,
					cabinName,
					outboundFlights[0].Time, outboundFlights[0].Seats,
					inboundFlights[0].Time, inboundFlights[0].Seats,
				)
			}
		}
	}

	return writer.Flush()
}

func filterSeats(flights []gavios.Flight, minSeats int) []gavios.Flight {
	return lo.Filter(flights, func(flight gavios.Flight, _ int) bool {
		return flight.Seats >= minSeats
	})
}

// flightsOnDate returns flights departing on the given YYYY-MM-DD date.
func flightsOnDate(flights []gavios.Flight, date string) []gavios.Flight {
	var matches []gavios.Flight

	for _, flight := range flights {
		if strings.HasPrefix(flight.Date, date) {
			matches = append(matches, flight)
		}
	}

	return matches
}
