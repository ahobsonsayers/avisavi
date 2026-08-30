package gavios

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/samber/lo"
)

const departureTimeLayout = "2006-01-02T15:04:05"

// Flight is a single flight with reward seats available.
type Flight struct {
	// Departure is the departure date and time.
	Departure time.Time `json:"-"`
	// Time is the departure time in HH:MM format.
	Time string `json:"time"`
	// Seats is the number of reward seats available.
	Seats int `json:"seats"`
	// Carrier is the airline code operating the flight (e.g. "BA").
	Carrier string `json:"carrier"`
}

func (flight *Flight) UnmarshalJSON(data []byte) error {
	type flightResponse struct {
		Date    string `json:"date"`
		Time    string `json:"time"`
		Seats   int    `json:"seats"`
		Carrier string `json:"carrier"`
	}

	var response flightResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	departure, err := time.Parse(departureTimeLayout, response.Date)
	if err != nil {
		return fmt.Errorf("parsing flight date %q: %w", response.Date, err)
	}

	flight.Departure = departure
	flight.Time = response.Time
	flight.Seats = response.Seats
	flight.Carrier = response.Carrier

	return nil
}

// TripFlights holds the flights for a trip, split by travel direction.
type TripFlights struct {
	// Outbound holds the outbound flights, ordered by departure date.
	Outbound []Flight
	// Inbound holds the inbound flights, ordered by departure date.
	Inbound []Flight
}

func (t TripFlights) FilterByDates(outboundDate, returnDate DateRange) TripFlights {
	if outboundDate.IsZero() && returnDate.IsZero() {
		return t
	}

	return TripFlights{
		Outbound: filterFlightsByDate(t.Outbound, outboundDate),
		Inbound:  filterFlightsByDate(t.Inbound, returnDate),
	}
}

func (t *TripFlights) UnmarshalJSON(data []byte) error {
	type flightsResponse struct {
		Flights map[string][]Flight `json:"flightsPerDate"`
	}

	type tripFlightsResponse struct {
		Outbound flightsResponse `json:"outbound"`
		Inbound  flightsResponse `json:"inbound"`
	}

	var response tripFlightsResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	t.Outbound = flightMapToSlice(response.Outbound.Flights)
	t.Inbound = flightMapToSlice(response.Inbound.Flights)

	return nil
}

// RouteFlights describes the reward flights on a route, grouped by cabin class.
type RouteFlights struct {
	Economy  TripFlights
	Premium  TripFlights
	Business TripFlights
	First    TripFlights
}

func (r *RouteFlights) UnmarshalJSON(data []byte) error {
	type availabilityPerCabinResponse struct {
		Economy  TripFlights `json:"Economy"`
		Premium  TripFlights `json:"Premium"`
		Business TripFlights `json:"Business"`
		First    TripFlights `json:"First"`
	}

	type routeFlightsResponse struct {
		AvailabilityPerCabin availabilityPerCabinResponse `json:"availabilityPerCabin"`
	}

	var response routeFlightsResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	cabins := response.AvailabilityPerCabin
	r.Economy = cabins.Economy
	r.Premium = cabins.Premium
	r.Business = cabins.Business
	r.First = cabins.First

	return nil
}

// flightMapToSlice converts a date -> flights map to a slice ordered by departure time.
func flightMapToSlice(flightMap map[string][]Flight) []Flight {
	var flights []Flight

	for _, perDate := range flightMap {
		flights = append(flights, perDate...)
	}

	slices.SortFunc(
		flights,
		func(left, right Flight) int {
			return left.Departure.Compare(right.Departure)
		},
	)

	return flights
}

func (r RouteFlights) FilterByDates(outboundDate, returnDate DateRange) RouteFlights {
	if outboundDate.IsZero() && returnDate.IsZero() {
		return r
	}

	return RouteFlights{
		Economy:  r.Economy.FilterByDates(outboundDate, returnDate),
		Premium:  r.Premium.FilterByDates(outboundDate, returnDate),
		Business: r.Business.FilterByDates(outboundDate, returnDate),
		First:    r.First.FilterByDates(outboundDate, returnDate),
	}
}

func filterFlightsByDate(flights []Flight, dateRange DateRange) []Flight {
	if dateRange.IsZero() {
		return flights
	}

	return lo.Filter(
		flights,
		func(flight Flight, _ int) bool {
			return dateRange.InRange(flight.Departure)
		},
	)
}
