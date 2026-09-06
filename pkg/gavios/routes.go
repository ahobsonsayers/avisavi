package gavios

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

// Network is the full route graph of airports and the routes between them.
type Network struct {
	// Airport code -> Airport details
	Airports map[string]Airport

	// Origin airport code -> Destination airport code -> Route details
	RouteMap map[string]map[string]RouteDetails
}

// Airport is an airport with its location metadata.
type Airport struct {
	AirportCode string `json:"airportCode"`
	AirportName string `json:"airportName"`
	CountryCode string `json:"countryCode,omitempty"`
	Country     string `json:"country"`
	City        string `json:"city"`
}

// Route is a route between two airports, with the route details.
type Route struct {
	Origin      Airport      `json:"origin"`
	Destination Airport      `json:"destination"`
	Details     RouteDetails `json:"details"`
}

// RouteDetails is everything known about a route, excluding its airports.
type RouteDetails struct {
	Region      string      `json:"region"`
	FlownBy     []string    `json:"flownBy"`
	AviosPrices AviosPrices `json:"aviosPrices"`
}

// AviosPrices gives the Avios price range per cabin class.
type AviosPrices struct {
	Economy  AviosPrice `json:"economy"`
	Premium  AviosPrice `json:"premium"`
	Business AviosPrice `json:"business"`
	First    AviosPrice `json:"first"`
}

// AviosPrice is the min and max Avios for one cabin class on a route.
type AviosPrice struct {
	MinAvios int `json:"minAvios"`
	MaxAvios int `json:"maxAvios"`
}

// Routes returns all routes in the network, ordered by origin then destination code.
func (r Network) Routes() []Route {
	routes := make([]Route, 0, len(r.RouteMap))

	originCodes := slices.Sorted(maps.Keys(r.RouteMap))
	for _, originCode := range originCodes {
		originRoutes := r.RouteMap[originCode]

		destinationCodes := slices.Sorted(maps.Keys(originRoutes))
		for _, destinationCode := range destinationCodes {
			routes = append(routes, Route{
				Origin:      r.Airports[originCode],
				Destination: r.Airports[destinationCode],
				Details:     originRoutes[destinationCode],
			})
		}
	}

	return routes
}

// AirportCodes returns all airport codes in the network.
func (r Network) AirportCodes() []string {
	return slices.Collect(maps.Keys(r.Airports))
}

// Regions returns geographic regions across all routes, sorted alphabetically.
func (r Network) Regions() []string {
	allRoutes := r.Routes()
	regions := mapset.NewSet[string]()
	for _, route := range allRoutes {
		if route.Details.Region != "" {
			regions.Add(route.Details.Region)
		}
	}

	regionsSlice := regions.ToSlice()
	slices.Sort(regionsSlice)
	return regionsSlice
}

// FindRoutesInput filters the routes returned by FindRoutes.
// A zero-value input returns all routes.
type FindRoutesInput struct {
	// If set finds flights from these origins.
	// If empty finds flights from any origin.
	Origins []string

	// If set finds flights to these destinations.
	// If empty finds flights to any destination.
	Destinations []string

	// If set finds flights to destinations in these regions.
	// If empty finds flights to any region.
	DestinationRegions []string
}

// FindRoutes returns the routes matching the input, ordered by origin then
// destination code. A zero-value input returns all routes.
func (r Network) FindRoutes(input FindRoutesInput) ([]Route, error) {
	routes := r.Routes()

	routes, err := r.filterRoutesByOriginDestination(routes, input.Origins, input.Destinations)
	if err != nil {
		return nil, err
	}

	if len(routes) == 0 {
		return nil, errors.New(
			"no routes from specified origins to destination",
		)
	}

	routes, err = r.filterRoutesByRegions(routes, input.DestinationRegions)
	if err != nil {
		return nil, err
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf(
			"no routes to specified destination regions: %s",
			strings.Join(input.DestinationRegions, ", "),
		)
	}

	return routes, nil
}

func (r Network) filterRoutesByOriginDestination(routes []Route, origins, destinations []string) ([]Route, error) {
	origins, err := NormalizeAirportCodes(origins, r.AirportCodes())
	if err != nil {
		return nil, err
	}

	destinations, err = NormalizeAirportCodes(destinations, r.AirportCodes())
	if err != nil {
		return nil, err
	}

	originSet := mapset.NewSet(origins...)
	destinationSet := mapset.NewSet(destinations...)

	filteredRoutes := make([]Route, 0, len(routes))
	for _, route := range routes {

		if originSet.Cardinality() > 0 && !originSet.Contains(route.Origin.AirportCode) {
			continue
		}

		if destinationSet.Cardinality() > 0 && !destinationSet.Contains(route.Destination.AirportCode) {
			continue
		}

		filteredRoutes = append(filteredRoutes, route)
	}

	return slices.Clip(filteredRoutes), nil
}

func (r Network) filterRoutesByRegions(routes []Route, wantedRegions []string) ([]Route, error) {
	validRegions := r.Regions()

	wantedRegions, err := NormalizeRegions(wantedRegions, validRegions)
	if err != nil {
		return nil, err
	}

	if len(wantedRegions) == 0 {
		return routes, nil
	}

	wantedRegionSet := mapset.NewSet(wantedRegions...)

	// Filter by regions
	regionRoutes := make([]Route, 0, len(routes))
	for _, route := range routes {
		routeRegion := strings.ToLower(route.Details.Region)
		if wantedRegionSet.Contains(routeRegion) {
			regionRoutes = append(regionRoutes, route)
		}
	}

	return slices.Clip(regionRoutes), nil
}
