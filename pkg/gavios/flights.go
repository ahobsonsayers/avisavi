package gavios

import (
	"time"

	"github.com/samber/lo"
)

// Flight is a single flight with reward seats available.
type Flight struct {
	// Departure is the departure date and time.
	Departure time.Time `json:"departure"`
	// Time is the departure time in HH:MM format.
	Time string `json:"time"`
	// Seats is the number of reward seats available.
	Seats int `json:"seats"`
	// Carrier is the airline code operating the flight (e.g. "BA").
	Carrier string `json:"carrier"`
}

// TripFlights holds the flights of a trip split by direction.
type TripFlights struct {
	// Outbound holds the outbound flights, ordered by departure date.
	Outbound []Flight `json:"outbound"`
	// Inbound holds the inbound flights, ordered by departure date.
	Inbound []Flight `json:"inbound"`
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

// RouteFlights describes the reward flights on a route per cabin class.
type RouteFlights struct {
	Economy  TripFlights `json:"economy"`
	Premium  TripFlights `json:"premium"`
	Business TripFlights `json:"business"`
	First    TripFlights `json:"first"`
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
