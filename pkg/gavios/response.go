package gavios

import (
	"fmt"
	"slices"
	"time"
)

// This file holds the private response types mirroring the avios api json

const departureTimeLayout = "2006-01-02T15:04:05"

type routesResponse struct {
	Origins []routeOriginResponse `json:"origins"`
}

func (r routesResponse) toNetwork() (Network, error) {
	network := Network{
		Airports: make(map[string]Airport),
		RouteMap: make(map[string]map[string]RouteDetails),
	}

	for _, origin := range r.Origins {
		originAirport, err := origin.toAirport()
		if err != nil {
			return Network{}, err
		}
		network.Airports[originAirport.AirportCode] = originAirport

		originRoutes := make(map[string]RouteDetails, len(origin.Destinations))
		for _, destination := range origin.Destinations {
			destinationAirport, err := destination.toAirport()
			if err != nil {
				return Network{}, err
			}
			network.Airports[destinationAirport.AirportCode] = destinationAirport
			originRoutes[destinationAirport.AirportCode] = destination.toRouteDetails()
		}

		network.RouteMap[originAirport.AirportCode] = originRoutes
	}

	return network, nil
}

type routeOriginResponse struct {
	AirportCode  string                     `json:"airportCode"`
	AirportName  string                     `json:"airportName"`
	CountryCode  string                     `json:"countryCode"`
	Country      string                     `json:"countryName"`
	City         string                     `json:"name"`
	Destinations []routeDestinationResponse `json:"destinations"`
}

func (r routeOriginResponse) toAirport() (Airport, error) {
	airportCode, err := NormalizeAirportCode(r.AirportCode, nil)
	if err != nil {
		return Airport{}, err
	}

	return Airport{
		AirportCode: airportCode,
		AirportName: r.AirportName,
		CountryCode: r.CountryCode,
		Country:     r.Country,
		City:        r.City,
	}, nil
}

type routeDestinationResponse struct {
	AirportCode string              `json:"airportCode"`
	AirportName string              `json:"airportName"`
	CountryCode string              `json:"countryCode"`
	Country     string              `json:"countryName"`
	City        string              `json:"name"`
	Region      []string            `json:"broadSearchCategories"`
	FlownBy     []string            `json:"flownByPartners"`
	AviosPrices aviosPricesResponse `json:"aviosPerCabinClass"`
}

func (r routeDestinationResponse) toAirport() (Airport, error) {
	airportCode, err := NormalizeAirportCode(r.AirportCode, nil)
	if err != nil {
		return Airport{}, err
	}

	return Airport{
		AirportCode: airportCode,
		AirportName: r.AirportName,
		CountryCode: r.CountryCode,
		Country:     r.Country,
		City:        r.City,
	}, nil
}

func (r routeDestinationResponse) toRouteDetails() RouteDetails {
	region := ""
	if len(r.Region) > 0 {
		region = r.Region[0]
	}

	return RouteDetails{
		Region:      region,
		FlownBy:     r.FlownBy,
		AviosPrices: r.AviosPrices.toAviosPrices(),
	}
}

type aviosPricesResponse struct {
	Economy  aviosPriceResponse `json:"Economy"`
	Premium  aviosPriceResponse `json:"Premium"`
	Business aviosPriceResponse `json:"Business"`
	First    aviosPriceResponse `json:"First"`
}

func (r aviosPricesResponse) toAviosPrices() AviosPrices {
	return AviosPrices{
		Economy:  r.Economy.toAviosPrice(),
		Premium:  r.Premium.toAviosPrice(),
		Business: r.Business.toAviosPrice(),
		First:    r.First.toAviosPrice(),
	}
}

type aviosPriceResponse struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func (r aviosPriceResponse) toAviosPrice() AviosPrice {
	return AviosPrice{MinAvios: r.Min, MaxAvios: r.Max}
}

type routeFlightsResponse struct {
	AvailabilityPerCabin availabilityResponse `json:"availabilityPerCabin"`
}

func (r routeFlightsResponse) toRouteFlights() (RouteFlights, error) {
	return r.AvailabilityPerCabin.toRouteFlights()
}

type availabilityResponse struct {
	Economy  tripFlightsResponse `json:"Economy"`
	Premium  tripFlightsResponse `json:"Premium"`
	Business tripFlightsResponse `json:"Business"`
	First    tripFlightsResponse `json:"First"`
}

func (r availabilityResponse) toRouteFlights() (RouteFlights, error) {
	economy, err := r.Economy.toTripFlights()
	if err != nil {
		return RouteFlights{}, err
	}

	premium, err := r.Premium.toTripFlights()
	if err != nil {
		return RouteFlights{}, err
	}

	business, err := r.Business.toTripFlights()
	if err != nil {
		return RouteFlights{}, err
	}

	first, err := r.First.toTripFlights()
	if err != nil {
		return RouteFlights{}, err
	}

	return RouteFlights{
		Economy:  economy,
		Premium:  premium,
		Business: business,
		First:    first,
	}, nil
}

type tripFlightsResponse struct {
	Outbound flightsPerDateResponse `json:"outbound"`
	Inbound  flightsPerDateResponse `json:"inbound"`
}

func (r tripFlightsResponse) toTripFlights() (TripFlights, error) {
	outbound, err := r.Outbound.toFlights()
	if err != nil {
		return TripFlights{}, err
	}

	inbound, err := r.Inbound.toFlights()
	if err != nil {
		return TripFlights{}, err
	}

	return TripFlights{Outbound: outbound, Inbound: inbound}, nil
}

type flightsPerDateResponse struct {
	Flights map[string][]flightResponse `json:"flightsPerDate"`
}

func (r flightsPerDateResponse) toFlights() ([]Flight, error) {
	flights := make([]Flight, 0, len(r.Flights))

	for _, perDate := range r.Flights {
		for _, flight := range perDate {
			converted, err := flight.toFlight()
			if err != nil {
				return nil, err
			}

			flights = append(flights, converted)
		}
	}

	slices.SortFunc(flights, func(left, right Flight) int {
		return left.Departure.Compare(right.Departure)
	})

	return flights, nil
}

type flightResponse struct {
	Date    string `json:"date"`
	Time    string `json:"time"`
	Seats   int    `json:"seats"`
	Carrier string `json:"carrier"`
}

func (r flightResponse) toFlight() (Flight, error) {
	departure, err := time.Parse(departureTimeLayout, r.Date)
	if err != nil {
		return Flight{}, fmt.Errorf("parsing flight date %q: %w", r.Date, err)
	}

	return Flight{
		Departure: departure,
		Time:      r.Time,
		Seats:     r.Seats,
		Carrier:   r.Carrier,
	}, nil
}
