package gavios

import (
	"context"
	"encoding/json"
	"maps"
	"net/url"
	"slices"
	"strconv"
)

type Routes struct {
	// Airport code -> Airport details
	Airports map[string]Airport

	// Origin airport code -> Destination airport code -> Route details
	Routes map[string]map[string]Route
}

type Airport struct {
	Code        string `json:"airportCode"`
	Name        string `json:"airportName"`
	CountryCode string `json:"countryCode,omitempty"`
	Country     string `json:"countryName"`
	City        string `json:"name"`
}

type Route struct {
	Region      string
	FlownBy     []string    `json:"flownByPartners"`
	AviosPrices AviosPrices `json:"aviosPerCabinClass"`
}

type AviosPrices struct {
	Economy  AviosPrice `json:"Economy"`
	Premium  AviosPrice `json:"Premium"`
	Business AviosPrice `json:"Business"`
	First    AviosPrice `json:"First"`
}

type AviosPrice struct {
	MinAvios int `json:"min"`
	MaxAvios int `json:"max"`
}

// Regions returns geographic regions across all routes, sorted alphabetically.
func (routes Routes) Regions() []string {
	regions := make(map[string]struct{})
	for _, originRoutes := range routes.Routes {
		for _, route := range originRoutes {
			if route.Region != "" {
				regions[route.Region] = struct{}{}
			}
		}
	}

	return slices.Sorted(maps.Keys(regions))
}

func (routes *Routes) UnmarshalJSON(data []byte) error {
	type originResponse struct {
		Destinations []json.RawMessage `json:"destinations"`
	}

	type routesResponse struct {
		Origins []json.RawMessage `json:"origins"`
	}

	var response routesResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	routes.Airports = make(map[string]Airport)
	routes.Routes = make(map[string]map[string]Route)

	for _, originRaw := range response.Origins {

		var origin originResponse
		err = json.Unmarshal(originRaw, &origin)
		if err != nil {
			return err
		}

		var originAirport Airport
		err = json.Unmarshal(originRaw, &originAirport)
		if err != nil {
			return err
		}

		routes.Airports[originAirport.Code] = originAirport

		originRoutes := make(map[string]Route)
		for _, destinationRaw := range origin.Destinations {
			var destinationAirport Airport
			err = json.Unmarshal(destinationRaw, &destinationAirport)
			if err != nil {
				return err
			}

			var route Route
			err = json.Unmarshal(destinationRaw, &route)
			if err != nil {
				return err
			}

			routes.Airports[destinationAirport.Code] = destinationAirport
			originRoutes[destinationAirport.Code] = route
		}

		routes.Routes[originAirport.Code] = originRoutes
	}

	return nil
}

func (airport *Airport) UnmarshalJSON(data []byte) error {
	type airportResponse struct {
		AirportCode string `json:"airportCode"`
		AirportName string `json:"airportName"`
		CountryCode string `json:"countryCode"`
		Country     string `json:"countryName"`
		City        string `json:"name"`
	}

	var response airportResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	airport.Code = response.AirportCode
	airport.Name = response.AirportName
	airport.CountryCode = response.CountryCode
	airport.Country = response.Country
	airport.City = response.City

	return nil
}

func (route *Route) UnmarshalJSON(data []byte) error {
	type routeResponse struct {
		BroadSearchCategories []string    `json:"broadSearchCategories"`
		Prices                AviosPrices `json:"aviosPerCabinClass"`
		FlownByPartners       []string    `json:"flownByPartners"`
	}

	var response routeResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	route.Region = ""
	if len(response.BroadSearchCategories) > 0 {
		route.Region = response.BroadSearchCategories[0]
	}
	route.AviosPrices = response.Prices
	route.FlownBy = response.FlownByPartners

	return nil
}

// Routes fetches reward destinations grouped by origin airport.
func (client *Client) Routes(ctx context.Context, adults int, oneWay bool) (Routes, error) {
	query := url.Values{}
	query.Set("ByAirport", "true")
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("OneWay", strconv.FormatBool(oneWay))

	var routes Routes
	err := client.get(
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
