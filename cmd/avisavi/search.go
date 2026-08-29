package main

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

var searchCmd = &cli.Command{
	Name:  "search",
	Usage: "Find destinations with seats on specific outbound and return dates",
	Description: `Search all reward destinations from an origin for available seats
on the given outbound and return dates.

This fetches the route list, then checks availability for each destination
(with a random 1-2s delay between requests to avoid rate limiting).

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
	destination  gavios.Destination
	availability gavios.Availability
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
		availabilities := make([]gavios.Availability, 0, len(results))
		for _, result := range results {
			availabilities = append(availabilities, result.availability)
		}
		return printJSON(availabilities)
	}

	return renderSearchResults(os.Stdout, results, outbound, returnDate, cabin, minSeats)
}

func searchDestinations(ctx context.Context, client *gavios.Client, origin string, adults int) []searchResult {
	routes, err := client.Routes(ctx, adults, false)
	if err != nil {
		return nil
	}

	filtered, err := routes.RoutesFromOrigin(origin)
	if err != nil {
		return nil
	}

	destinations := filtered.Origins[0].Destinations

	destinations = lo.UniqBy(
		destinations,
		func(destination gavios.Destination) string {
			return destination.DestinationAirportCode
		},
	)

	results := make([]searchResult, 0, len(destinations))
	for _, destination := range destinations {
		availability, err := client.Availability(
			ctx,
			origin,
			destination.DestinationAirportCode,
			false,
			adults,
		)
		if err != nil {
			continue
		}

		// Random sleep to prevent being blocked
		time.Sleep(time.Duration(1000+rand.IntN(1000)) * time.Millisecond)

		results = append(
			results,
			searchResult{
				destination:  destination,
				availability: availability,
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
		cabins := result.availability.CabinAvailability
		if cabin != "" {
			cabins = map[string]gavios.CabinData{}
			if cabinData, ok := result.availability.CabinAvailability[cabin]; ok {
				cabins[cabin] = cabinData
			}
		}

		for cabinName, cabinData := range cabins {
			outboundFlights := filterSeats(flightsOnDate(cabinData.Outbound.Flights, outbound), minSeats)
			inboundFlights := filterSeats(flightsOnDate(cabinData.Inbound.Flights, returnDate), minSeats)
			if len(outboundFlights) > 0 && len(inboundFlights) > 0 {
				fmt.Fprintf(
					writer, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\n",
					result.destination.DestinationAirportCode,
					result.destination.DestinationName,
					result.destination.CountryName,
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
func flightsOnDate(flightsPerDate map[string][]gavios.Flight, date string) []gavios.Flight {
	var matches []gavios.Flight
	for _, flights := range flightsPerDate {
		for _, flight := range flights {
			if strings.HasPrefix(flight.Date, date) {
				matches = append(matches, flight)
			}
		}
	}
	return matches
}
