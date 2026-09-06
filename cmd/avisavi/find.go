package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
	"github.com/urfave/cli/v3"
)

var findCmd = &cli.Command{
	Name:  "find",
	Usage: "Find reward flights across the route network",
	Description: `Find reward flights, filtered by origin, destination, region, cabin
and dates.

Prints detailed per-flight availability with colour-coded seat counts.

Examples:
  avisavi find --origin LON --outbound 2026-09-09 --return 2026-09-13
  avisavi find --origin LON --outbound 2026-09-09 --return 2026-09-13 --cabin Business
  avisavi find --origin LON --outbound 2026-09-09 --return 2026-09-13 --adults 2
  avisavi find --origin LON --destination JFK
  avisavi find --region "North America"`,
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
		&cli.StringFlag{Name: "outbound", Usage: "outbound date YYYY-MM-DD"},
		&cli.StringFlag{Name: "outbound-after", Usage: "outbound after date YYYY-MM-DD (exclusive)"},
		&cli.StringFlag{Name: "outbound-before", Usage: "outbound before date YYYY-MM-DD (exclusive)"},
		&cli.StringFlag{Name: "return", Usage: "return date YYYY-MM-DD"},
		&cli.StringFlag{Name: "return-after", Usage: "return after date YYYY-MM-DD (exclusive)"},
		&cli.StringFlag{Name: "return-before", Usage: "return before date YYYY-MM-DD (exclusive)"},
		&cli.IntFlag{Name: "adults", Aliases: []string{"a"}, Value: 1, Usage: "number of adults"},
		&cli.BoolFlag{Name: "one-way", Usage: "one-way flights only"},
		&cli.StringFlag{
			Name:  "cabin",
			Usage: "filter by cabin (Economy, Premium, Business, First)",
		},
		&cli.BoolFlag{Name: "json", Usage: "print raw JSON"},
	},
	Action: findAction,
}

func findAction(ctx context.Context, cmd *cli.Command) error {
	client, err := getAviosClient(ctx)
	if err != nil {
		return err
	}

	input, err := findInput(cmd)
	if err != nil {
		return err
	}

	found, err := client.FindFlights(ctx, input)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return printJSON(found)
	}

	cabin := cmd.String("cabin")

	return renderFlights(os.Stdout, found, cabin)
}

func findInput(cmd *cli.Command) (gavios.FindFlightsInput, error) {
	outbound, err := parseDateRange(cmd, "outbound")
	if err != nil {
		return gavios.FindFlightsInput{}, err
	}

	returnDateRange, err := parseDateRange(cmd, "return")
	if err != nil {
		return gavios.FindFlightsInput{}, err
	}

	return gavios.FindFlightsInput{
		Origins:            cmd.StringSlice("origin"),
		Destinations:       cmd.StringSlice("destination"),
		DestinationRegions: cmd.StringSlice("region"),
		Adults:             cmd.Int("adults"),
		OneWay:             cmd.Bool("one-way"),
		Outbound:           outbound,
		Return:             returnDateRange,
	}, nil
}

// parseDateRange parses the <direction>, -after and -before flags
// for the given direction into a DateRange.
func parseDateRange(cmd *cli.Command, direction string) (gavios.DateRange, error) {
	var dateRange gavios.DateRange

	on := cmd.String(direction)
	if on != "" {
		onDate, err := time.Parse("2006-01-02", on)
		if err != nil {
			return dateRange, fmt.Errorf("invalid %s date %q: %w", direction, on, err)
		}
		dateRange.On = onDate
	}

	after := cmd.String(direction + "-after")
	if after != "" {
		afterDate, err := time.Parse("2006-01-02", after)
		if err != nil {
			return dateRange, fmt.Errorf("invalid %s-after date %q: %w", direction, after, err)
		}
		dateRange.After = afterDate
	}

	before := cmd.String(direction + "-before")
	if before != "" {
		beforeDate, err := time.Parse("2006-01-02", before)
		if err != nil {
			return dateRange, fmt.Errorf("invalid %s-before date %q: %w", direction, before, err)
		}
		dateRange.Before = beforeDate
	}

	return dateRange, nil
}

// cabinFlights pairs a cabin name with its flights.
type cabinFlights struct {
	name        string
	tripFlights gavios.TripFlights
}

// cabinTripFlights pairs each cabin with its name, filtered by the wanted
// cabin (case-insensitive). All cabins are returned when the name is empty.
func cabinTripFlights(routeFlights gavios.RouteFlights, cabin string) []cabinFlights {
	cabins := []cabinFlights{
		{"Economy", routeFlights.Economy},
		{"Premium", routeFlights.Premium},
		{"Business", routeFlights.Business},
		{"First", routeFlights.First},
	}

	if cabin == "" {
		return cabins
	}

	for _, cabinFlight := range cabins {
		if strings.EqualFold(cabinFlight.name, cabin) {
			return []cabinFlights{cabinFlight}
		}
	}

	return nil
}

// renderFlights prints detailed per-flight availability for every route and
// cabin, with colour-coded seat counts.
func renderFlights(w io.Writer, found []gavios.FoundFlights, cabin string) error {
	for _, flights := range found {
		routeHeader := routeHeaderStyle.Render(fmt.Sprintf(
			"%s → %s",
			flights.Origin.AirportCode,
			flights.Destination.AirportCode,
		))
		fmt.Fprintf(w, "\n%s\n", routeHeader)

		cabins := cabinTripFlights(flights.RouteFlights, cabin)

		for _, cabinFlight := range cabins {
			if len(cabinFlight.tripFlights.Outbound) == 0 &&
				len(cabinFlight.tripFlights.Inbound) == 0 {
				continue
			}

			cabinHeader := cabinStyles[cabinFlight.name].Render(cabinFlight.name)
			fmt.Fprintf(w, "\n%s\n", cabinHeader)

			renderTripFlights(w, "Outbound", cabinFlight.tripFlights.Outbound)
			renderTripFlights(w, "Inbound", cabinFlight.tripFlights.Inbound)
		}
	}

	return nil
}

// renderTripFlights prints flights for one direction as a borderless table
// with aligned columns, colour-coded by seat count.
func renderTripFlights(w io.Writer, name string, flights []gavios.Flight) {
	if len(flights) == 0 {
		return
	}

	rows := make([][]string, 0, len(flights))
	for _, flight := range flights {
		rows = append(
			rows,
			[]string{
				flight.Departure.Format("Mon 02 Jan 2006"),
				flight.Time,
				seatStyle(flight.Seats).Render(fmt.Sprintf("%d seats", flight.Seats)),
				carrierStyle.Render(flight.Carrier),
			},
		)
	}

	flightTable := newFlightTable(rows)

	fmt.Fprintln(w, directionStyle.Render(name))
	fmt.Fprintln(w, flightTable.Render())
}
