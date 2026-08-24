package gavios

import (
	"context"
	"net/url"
	"strconv"
)

// Availability describes reward seats on a route, grouped by cabin class and direction.
type Availability struct {
	// CabinAvailability maps a cabin class (e.g. "Economy") to its flight data.
	CabinAvailability map[string]CabinData `json:"availabilityPerCabin"`
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
	// Flights maps a date to the flights with reward seats that day.
	Flights map[string][]Flight `json:"flightsPerDate"`
}

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

	var availability Availability
	err = c.get(
		ctx,
		"/spend/v1/flight/allcabins",
		query,
		&availability,
	)
	if err != nil {
		return Availability{}, err
	}

	return availability, nil
}
