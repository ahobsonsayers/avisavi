package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ahobsonsayers/avisavi/pkg/gavios"
	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
)

var findCmd = &cli.Command{
	Name:  "find",
	Usage: "Find reward flights across the route network",
	Description: `Find reward flights, filtered by origin, destination, region, cabin,
dates, and minimum seats.

Without dates, prints detailed per-flight availability with colour-coded
seat counts. With outbound and/or return dates, prints a round-trip
summary table.

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
		&cli.StringFlag{Name: "return", Usage: "return date YYYY-MM-DD"},
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

	hasDates := cmd.String("outbound") != "" || cmd.String("return") != ""
	if hasDates {
		return renderFlightsWithDates(os.Stdout, found, cabin)
	}

	return renderFlightsWithoutDates(os.Stdout, found, cabin)
}

func findInput(cmd *cli.Command) (gavios.FindFlightsInput, error) {
	input := gavios.FindFlightsInput{
		Origins:            cmd.StringSlice("origin"),
		Destinations:       cmd.StringSlice("destination"),
		DestinationRegions: cmd.StringSlice("region"),
		Adults:             cmd.Int("adults"),
		OneWay:             cmd.Bool("one-way"),
	}

	outbound := cmd.String("outbound")
	if outbound != "" {
		outboundDate, err := time.Parse("2006-01-02", outbound)
		if err != nil {
			return input, fmt.Errorf("invalid outbound date %q: %w", outbound, err)
		}
		input.Outbound = gavios.DateRange{On: outboundDate}
	}

	returnDate := cmd.String("return")
	if returnDate != "" {
		returnDateParsed, err := time.Parse("2006-01-02", returnDate)
		if err != nil {
			return input, fmt.Errorf("invalid return date %q: %w", returnDate, err)
		}
		input.Return = gavios.DateRange{On: returnDateParsed}
	}

	return input, nil
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

// renderFlightsWithDates prints a round-trip summary table: one row per
// cabin, showing the first outbound and inbound flights of each day range.
func renderFlightsWithDates(w io.Writer, found []gavios.FoundFlights, cabin string) error {
	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	fmt.Fprint(writer, "DEST\tNAME\tCOUNTRY\tCABIN\tOUT\tSEATS\tRET\tSEATS\n")

	for _, flights := range found {
		cabins := cabinTripFlights(flights.RouteFlights, cabin)

		for _, cabinFlight := range cabins {
			outboundFlights := cabinFlight.tripFlights.Outbound
			inboundFlights := cabinFlight.tripFlights.Inbound
			if len(outboundFlights) == 0 || len(inboundFlights) == 0 {
				continue
			}

			fmt.Fprintf(
				writer, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\n",
				flights.Destination.AirportCode,
				flights.Destination.City,
				flights.Destination.Country,
				cabinFlight.name,
				outboundFlights[0].Time, outboundFlights[0].Seats,
				inboundFlights[0].Time, inboundFlights[0].Seats,
			)
		}
	}

	return writer.Flush()
}

// renderFlightsWithoutDates prints detailed per-flight availability with
// colour-coded seat counts.
func renderFlightsWithoutDates(w io.Writer, found []gavios.FoundFlights, cabin string) error {
	for _, flights := range found {
		fmt.Fprintf(
			w, "\n=== %s → %s ===\n",
			flights.Origin.AirportCode,
			flights.Destination.AirportCode,
		)

		cabins := cabinTripFlights(flights.RouteFlights, cabin)

		for _, cabinFlight := range cabins {
			if len(cabinFlight.tripFlights.Outbound) == 0 &&
				len(cabinFlight.tripFlights.Inbound) == 0 {
				continue
			}

			fmt.Fprintf(w, "\n--- %s ---\n", cabinFlight.name)
			renderFlights(w, "Outbound", cabinFlight.tripFlights.Outbound)
			renderFlights(w, "Inbound", cabinFlight.tripFlights.Inbound)
		}
	}

	return nil
}

func renderFlights(w io.Writer, name string, flights []gavios.Flight) {
	if len(flights) == 0 {
		return
	}

	// Flights are already ordered by departure date after unmarshalling.
	fmt.Fprintf(w, "\n  %s:\n", name)

	for _, flight := range flights {
		line := fmt.Sprintf(
			"    %s %s seats=%d %s",
			flight.Departure.Format("2006-01-02"),
			flight.Time,
			flight.Seats,
			flight.Carrier,
		)
		fmt.Fprintln(w, seatLineColour(flight.Seats, line))
	}
}

var (
	seatGreen  = color.New(color.FgGreen).SprintFunc()
	seatYellow = color.New(color.FgYellow).SprintFunc()
	seatRed    = color.New(color.FgRed).SprintFunc()
)

// seatLineColour colours a flight line by seat availability. Colour is
// disabled automatically for non-TTY output or when NO_COLOR is set.
func seatLineColour(seats int, line string) string {
	switch {
	case seats >= 9:
		return seatGreen(line)
	case seats >= 5:
		return seatYellow(line)
	case seats >= 1:
		return seatRed(line)
	default:
		return line
	}
}
