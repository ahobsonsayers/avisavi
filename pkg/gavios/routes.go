package gavios

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Routes lists reward destinations available from one or more origins.
type Routes struct {
	// Origins is the set of origin airports and their bookable destinations.
	Origins []Origin `json:"origins"`
	// BroadSearchGroups groups destinations by continent/area with price ranges.
	BroadSearchGroups []Category `json:"broadSearchGroups"`
}

// Origin is a departure airport and every destination bookable from it.
type Origin struct {
	// AirportCode is the IATA code of the departure airport.
	AirportCode string `json:"airportCode"`
	// AirportName is the full name of the departure airport.
	AirportName string `json:"airportName"`
	// Name is the name of the city the airport serves.
	Name string `json:"name"`
	// CountryName is the full country name of the departure airport.
	CountryName string `json:"countryName"`
	// Destinations are the reward destinations reachable from this origin.
	Destinations []Destination `json:"destinations"`
}

// Destination is a reward destination reachable from an origin.
type Destination struct {
	// DestinationAirportCode is the IATA code of the destination airport.
	DestinationAirportCode string `json:"airportCode"`
	// DestinationAirportName is the full name of the destination airport.
	DestinationAirportName string `json:"airportName"`
	// DestinationName is the name of the city the airport serves.
	DestinationName string `json:"name"`
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

// Routes returns reward destinations grouped by origin airport.
func (c *Client) Routes(ctx context.Context, adults int, oneWay bool) (Routes, error) {
	query := url.Values{}
	query.Set("ByAirport", "true")
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("OneWay", strconv.FormatBool(oneWay))

	var routes Routes
	err := c.get(
		ctx,
		"/spend/v1/flight/routes",
		query,
		&routes,
	)
	if err != nil {
		return Routes{}, err
	}
	return routes, nil
}

// RoutesFromOrigin filters routes to a single origin airport.
// An empty origin returns all routes.
func (routes Routes) RoutesFromOrigin(originAirport string) (Routes, error) {
	if originAirport == "" {
		return routes, nil
	}

	origin, err := normalizeAirportCode(originAirport)
	if err != nil {
		return Routes{}, err
	}

	for _, o := range routes.Origins {
		if o.AirportCode == origin {
			return Routes{
				Origins:           []Origin{o},
				BroadSearchGroups: routes.BroadSearchGroups,
			}, nil
		}
	}

	return Routes{}, fmt.Errorf("origin %q not found in routes", origin)
}
