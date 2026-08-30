package gavios

import (
	"context"
	"encoding/json"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

// Flight is a single flight with reward seats available.
type Flight struct {
	// Date is the departure date and time (ISO timestamp).
	Date string `json:"date"`
	// Time is the departure time in HH:MM format.
	Time string `json:"time"`
	// Seats is the number of reward seats available.
	Seats int `json:"seats"`
	// Carrier is the airline code operating the flight (e.g. "BA").
	Carrier string `json:"carrier"`
}

// TripFlights holds the flights for a trip, split by travel direction.
type TripFlights struct {
	// Outbound holds the outbound flights, ordered by departure date.
	Outbound []Flight
	// Inbound holds the inbound flights, ordered by departure date.
	Inbound []Flight
}

func (tripFlights *TripFlights) UnmarshalJSON(data []byte) error {
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

	tripFlights.Outbound = flightMapToSlice(response.Outbound.Flights)
	tripFlights.Inbound = flightMapToSlice(response.Inbound.Flights)

	return nil
}

// RouteFlights describes the reward flights on a route, grouped by cabin class.
type RouteFlights struct {
	Economy  TripFlights
	Premium  TripFlights
	Business TripFlights
	First    TripFlights
}

func (routeFlights *RouteFlights) UnmarshalJSON(data []byte) error {
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
	routeFlights.Economy = cabins.Economy
	routeFlights.Premium = cabins.Premium
	routeFlights.Business = cabins.Business
	routeFlights.First = cabins.First

	return nil
}

func (client *Client) RouteFlights(
	ctx context.Context,
	origin, destination string,
	oneWay bool,
	adults int,
) (RouteFlights, error) {
	origin, err := NormalizeAirportCode(origin)
	if err != nil {
		return RouteFlights{}, err
	}

	destination, err = NormalizeAirportCode(destination)
	if err != nil {
		return RouteFlights{}, err
	}

	query := url.Values{}
	query.Set("Origin", origin)
	query.Set("Destination", destination)
	query.Set("OneWay", strconv.FormatBool(oneWay))
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("IncludeNonBookableFlights", "true")

	var routeFlights RouteFlights
	err = client.get(
		ctx,
		"/spend/v1/flight/allcabins",
		query,
		&routeFlights,
	)
	if err != nil {
		return RouteFlights{}, err
	}

	return routeFlights, nil
}

// flightMapToSlice converts a date -> flights map to a slice ordered by
// departure timestamp.
func flightMapToSlice(flightMap map[string][]Flight) []Flight {
	var flights []Flight

	for _, perDate := range flightMap {
		flights = append(flights, perDate...)
	}

	slices.SortFunc(
		flights,
		func(left, right Flight) int {
			return strings.Compare(left.Date, right.Date)
		},
	)

	return flights
}
