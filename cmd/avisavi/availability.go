package main

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
	"github.com/urfave/cli/v3"
)

var availabilityCmd = &cli.Command{
	Name:  "availability",
	Usage: "Check seat availability across all cabins for a route",
	Description: `Show available reward seats for a route across Economy, Premium,
Business, and First cabins. Seat counts are colour-coded:
  GREEN  >= 9 seats
  YELLOW >= 5 seats
  RED    >= 1 seat
  ---     0 seats

Examples:
  avisavi availability --origin LON --destination ABV
  avisavi availability --origin LON --destination NYC --adults 2
  avisavi availability --origin LON --destination JFK --one-way`,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "origin", Aliases: []string{"o"}, Required: true, Usage: "IATA origin code"},
		&cli.StringFlag{Name: "destination", Aliases: []string{"d"}, Required: true, Usage: "IATA destination code"},
		&cli.IntFlag{Name: "adults", Aliases: []string{"a"}, Value: 1, Usage: "number of adults"},
		&cli.BoolFlag{Name: "one-way", Usage: "one-way flights only"},
		&cli.BoolFlag{Name: "json", Usage: "print raw JSON"},
	},
	Action: availabilityAction,
}

func availabilityAction(ctx context.Context, cmd *cli.Command) error {
	client, err := getAviosClient(ctx)
	if err != nil {
		return err
	}

	origin := cmd.String("origin")
	destination := cmd.String("destination")
	availability, err := client.Availability(
		ctx,
		origin,
		destination,
		cmd.Bool("one-way"),
		cmd.Int("adults"),
	)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return printJSON(availability)
	}

	return renderAvailability(os.Stdout, availability)
}

func renderAvailability(w io.Writer, availability gavios.Availability) error {
	for cabin, cabinData := range availability.CabinAvailability {
		fmt.Fprintf(w, "\n=== %s ===\n", cabin)
		renderDirection(w, "outbound", cabinData.Outbound)
		renderDirection(w, "inbound", cabinData.Inbound)
	}

	return nil
}

func renderDirection(w io.Writer, name string, direction gavios.Direction) {
	if len(direction.Flights) == 0 {
		return
	}

	// Iterate dates in order so output is stable and chronological.
	fmt.Fprintf(w, "\n  %s:\n", name)

	dates := maps.Keys(direction.Flights)
	sortedDates := slices.Sorted(dates)
	for _, date := range sortedDates {
		for _, flight := range direction.Flights[date] {
			fmt.Fprintf(
				w, "    %s %s seats=%d %s [%s]\n",
				flight.Date[:10],
				flight.Time,
				flight.Seats,
				flight.Carrier,
				seatColour(flight.Seats),
			)
		}
	}
}

// Seat availability colour bands (the API no longer reports thresholds).
func seatColour(seats int) string {
	switch {
	case seats >= 9:
		return "GREEN"
	case seats >= 5:
		return "YELLOW"
	case seats >= 1:
		return "RED"
	default:
		return "---"
	}
}
