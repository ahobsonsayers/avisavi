package gavios

import (
	"context"
	"net/url"
	"strconv"
)

// Routes lists reward destinations available from one or more origins.
type Routes struct {
	// Origins is the set of origin airports and their bookable destinations.
	Origins []Origin `json:"origins"`
	// BroadSearchCategories groups destinations by high-level theme (e.g. city, beach).
	BroadSearchCategories []Category `json:"broadSearchCategories"`
}

// Origin is a departure airport and every destination bookable from it.
type Origin struct {
	// OriginAirportCode is the IATA code of the departure airport.
	OriginAirportCode string `json:"originAirportCode"`
	// Destinations are the reward destinations reachable from this origin.
	Destinations []Destination `json:"destinations"`
}

// Destination is a reward destination reachable from an origin.
type Destination struct {
	// DestinationAirportCode is the IATA code of the destination airport.
	DestinationAirportCode string `json:"destinationAirportCode"`
	// DestinationAirportName is the full name of the destination airport.
	DestinationAirportName string `json:"destinationAirportName"`
	// DestinationName is the name of the city the airport serves.
	DestinationName string `json:"destinationName"`
	// CountryCode is the ISO country code of the destination.
	CountryCode string `json:"countryCode"`
	// CountryName is the full country name of the destination.
	CountryName string `json:"countryName"`
	// BroadSearchCategories lists the theme categories this destination belongs to.
	BroadSearchCategories []string `json:"broadSearchCategories"`
	// AviosPerCabinClass maps a cabin class (e.g. "Economy") to its price range.
	AviosPerCabinClass map[string]CabinPrice `json:"aviosPerCabinClass"`
	// FlownByPartners lists partner airlines that operate reward flights here.
	FlownByPartners []string `json:"flownByPartners"`
}

// CabinPrice is the Avios price range for a single cabin class.
type CabinPrice struct {
	// Min is the lowest Avios price for this cabin class.
	Min int `json:"min"`
	// Max is the highest Avios price for this cabin class.
	Max int `json:"max"`
}

// Category groups destinations under a single high-level theme.
type Category struct {
	// Name is the theme name (e.g. "Beach", "City break").
	Name string `json:"name"`
	// DestinationCount is the number of destinations in this category.
	DestinationCount int `json:"destinationCount"`
	// AviosPerCabinClass maps a cabin class to its price range.
	AviosPerCabinClass map[string]CabinPrice `json:"aviosPerCabinClass"`
}

func (c *Client) Routes(ctx context.Context, originAirport string, adults int, oneWay bool) (Routes, error) {
	query := url.Values{}
	query.Set("ByAirport", "true")
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("OneWay", strconv.FormatBool(oneWay))
	if originAirport != "" {
		query.Set("OriginAirport", originAirport)
	}

	var routes Routes
	err := c.get(
		ctx,
		"/spend/v3/programmes/BAEC/GB/flight/routes",
		query,
		authRaw,
		&routes,
	)
	if err != nil {
		return Routes{}, err
	}
	return routes, nil
}
