package gavios

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

// RouteNetwork is the full route graph of airports and the routes between them.
type RouteNetwork struct {
	// Airport code -> Airport details
	Airports map[string]Airport

	// Origin airport code -> Destination airport code -> Route details
	Routes map[string]map[string]RouteDetails
}

// Airport is an airport with its location metadata.
type Airport struct {
	AirportCode string `json:"airportCode"`
	AirportName string `json:"airportName"`
	CountryCode string `json:"countryCode,omitempty"`
	Country     string `json:"countryName"`
	City        string `json:"name"`
}

// RouteDetails is everything known about a route, excluding its airports.
type RouteDetails struct {
	Region      string
	FlownBy     []string    `json:"flownByPartners"`
	AviosPrices AviosPrices `json:"aviosPerCabinClass"`
}

// AviosPrices gives the Avios price range per cabin class.
type AviosPrices struct {
	Economy  AviosPrice `json:"Economy"`
	Premium  AviosPrice `json:"Premium"`
	Business AviosPrice `json:"Business"`
	First    AviosPrice `json:"First"`
}

// AviosPrice is the min and max Avios for one cabin class on a route.
type AviosPrice struct {
	MinAvios int `json:"min"`
	MaxAvios int `json:"max"`
}

// Route is a route between two airports, with the route details.
type Route struct {
	Origin      Airport
	Destination Airport
	Details     RouteDetails
}

// GetRoute returns the route between two airports, or an error if none exists.
func (r RouteNetwork) GetRoute(originCode, destinationCode string) (Route, error) {
	originCode, err := NormalizeAirportCode(originCode)
	if err != nil {
		return Route{}, err
	}

	destinationCode, err = NormalizeAirportCode(destinationCode)
	if err != nil {
		return Route{}, err
	}

	details, found := r.Routes[originCode][destinationCode]
	if !found {
		if _, originFound := r.Airports[originCode]; !originFound {
			return Route{}, fmt.Errorf("origin %q not found in routes", originCode)
		}
		return Route{}, fmt.Errorf("no route %q -> %q", originCode, destinationCode)
	}

	return Route{
		Origin:      r.Airports[originCode],
		Destination: r.Airports[destinationCode],
		Details:     details,
	}, nil
}

// GetRoutes returns routes from one origin (empty code for all origins), sorted by destination.
func (r RouteNetwork) GetRoutes(originCode string) ([]Route, error) {
	if originCode == "" {
		return r.GetAllRoutes()
	}

	originCode, err := NormalizeAirportCode(originCode)
	if err != nil {
		return nil, err
	}

	destinationRoutes, found := r.Routes[originCode]
	if !found {
		return nil, fmt.Errorf("origin %q not found in routes", originCode)
	}

	routeList := make([]Route, 0, len(destinationRoutes))
	for _, destinationCode := range slices.Sorted(maps.Keys(destinationRoutes)) {
		routeList = append(routeList, Route{
			Origin:      r.Airports[originCode],
			Destination: r.Airports[destinationCode],
			Details:     destinationRoutes[destinationCode],
		})
	}

	return routeList, nil
}

// GetAllRoutes returns every route in the network, sorted by origin then destination.
func (r RouteNetwork) GetAllRoutes() ([]Route, error) {
	routeList := make([]Route, 0)
	for _, origin := range slices.Sorted(maps.Keys(r.Routes)) {
		originRoutes, err := r.GetRoutes(origin)
		if err != nil {
			return nil, err
		}
		routeList = append(routeList, originRoutes...)
	}
	return routeList, nil
}

// Regions returns geographic regions across all routes, sorted alphabetically.
func (r RouteNetwork) Regions() []string {
	regions := make(map[string]struct{})
	for _, originRoutes := range r.Routes {
		for _, details := range originRoutes {
			if details.Region != "" {
				regions[details.Region] = struct{}{}
			}
		}
	}

	return slices.Sorted(maps.Keys(regions))
}

func (r *RouteNetwork) UnmarshalJSON(data []byte) error {
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

	r.Airports = make(map[string]Airport)
	r.Routes = make(map[string]map[string]RouteDetails)

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

		r.Airports[originAirport.AirportCode] = originAirport

		originRoutes := make(map[string]RouteDetails)
		for _, destinationRaw := range origin.Destinations {
			var destinationAirport Airport
			err = json.Unmarshal(destinationRaw, &destinationAirport)
			if err != nil {
				return err
			}

			var route RouteDetails
			err = json.Unmarshal(destinationRaw, &route)
			if err != nil {
				return err
			}

			r.Airports[destinationAirport.AirportCode] = destinationAirport
			originRoutes[destinationAirport.AirportCode] = route
		}

		r.Routes[originAirport.AirportCode] = originRoutes
	}

	return nil
}

func (a *Airport) UnmarshalJSON(data []byte) error {
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

	a.AirportCode = response.AirportCode
	a.AirportName = response.AirportName
	a.CountryCode = response.CountryCode
	a.Country = response.Country
	a.City = response.City

	return nil
}

func (r *RouteDetails) UnmarshalJSON(data []byte) error {
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

	r.Region = ""
	if len(response.BroadSearchCategories) > 0 {
		r.Region = response.BroadSearchCategories[0]
	}
	r.AviosPrices = response.Prices
	r.FlownBy = response.FlownByPartners

	return nil
}
