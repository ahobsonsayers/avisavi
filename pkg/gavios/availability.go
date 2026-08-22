package gavios

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Availability describes reward seats on a route, grouped by cabin class and direction.
type Availability struct {
	// Origin is the departure airport.
	Origin LocationCode `json:"origin"`
	// Destination is the arrival airport.
	Destination LocationCode `json:"destination"`
	// CabinAvailability maps a cabin class (e.g. "Economy") to its flight data.
	CabinAvailability map[string]CabinData `json:"cabinAvailability"`
	// HighSeatAvailabilityThreshold is the seat count at/above which a flight is "high availability".
	HighSeatAvailabilityThreshold int `json:"highSeatAvailabilityThreshold"`
	// MediumAvailabilityThreshold is the seat count at/above which a flight is "medium availability".
	MediumAvailabilityThreshold int `json:"mediumAvailabilityThreshold"`
	// LowSeatAvailabilityThreshold is the seat count at/above which a flight is "low availability".
	LowSeatAvailabilityThreshold int `json:"lowSeatAvailabilityThreshold"`
}

// LocationCode is an airport identified by code and name.
type LocationCode struct {
	// Name is the full airport name.
	Name string `json:"name"`
	// Code is the IATA airport code.
	Code string `json:"code"`
}

// CabinData holds availability for a single cabin class across both travel directions.
type CabinData struct {
	// Outbound is the outbound direction (origin to destination).
	Outbound Direction `json:"outbound"`
	// Inbound is the inbound direction (destination to origin).
	Inbound Direction `json:"inbound"`
}

// Direction is the reward availability for a single direction of travel.
type Direction struct {
	// FromAvios is the minimum Avios price to book this direction.
	FromAvios int `json:"fromAvios"`
	// AvailabilityFrom is the earliest date reward seats are available (ISO timestamp).
	AvailabilityFrom string `json:"availabilityFrom"`
	// AvailabilityTo is the latest date reward seats are available (ISO timestamp).
	AvailabilityTo string `json:"availabilityTo"`
	// AvailableFlights maps a date to the flights with reward seats that day.
	AvailableFlights map[string][]Flight `json:"availableFlights"`
}

// Flight is a single flight with reward seats available.
type Flight struct {
	// Date is the departure date and time (ISO timestamp).
	Date string `json:"date"`
	// Time is the departure time in HH:MM format.
	Time string `json:"time"`
	// Peak reports whether the flight uses peak pricing.
	Peak bool `json:"peak"`
	// Direct reports whether the flight is non-stop.
	Direct bool `json:"direct"`
	// Avios is the reward price of this flight in Avios.
	Avios int `json:"avios"`
	// Seats is the number of reward seats available.
	Seats int `json:"seats"`
	// Carrier is the airline code operating the flight (e.g. "BA").
	Carrier string `json:"carrier"`
}

func (c *Client) Availability(
	ctx context.Context,
	origin, destination string,
	oneWay bool,
	adults int,
) (Availability, error) {
	origin, err := normalizeAirportCode(origin)
	if err != nil {
		return Availability{}, err
	}

	destination, err = normalizeAirportCode(destination)
	if err != nil {
		return Availability{}, err
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

	membershipNumber, err := c.MembershipNumber()
	if err != nil {
		return Availability{}, err
	}

	var availability Availability
	err = c.get(
		ctx,
		fmt.Sprintf("/spend/v3/programmes/BAEC/GB/%s/flight/availability/allcabins", membershipNumber),
		query, authRaw,
		&availability,
	)
	if err != nil {
		return Availability{}, err
	}

	return availability, nil
}
