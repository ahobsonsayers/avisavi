package gavios

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
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

// FindRoutes returns the routes matching the input, sorted by origin then
// destination code. A zero-value input returns all routes.
func (r RouteNetwork) FindRoutes(input FindRoutesInput) ([]Route, error) {
	origins, err := NormalizeAirportCodes(input.Origins)
	if err != nil {
		return nil, err
	}

	destinations, err := NormalizeAirportCodes(input.Destinations)
	if err != nil {
		return nil, err
	}

	for _, originCode := range origins {
		if _, found := r.Routes[originCode]; !found {
			return nil, fmt.Errorf("origin %q not found in routes", originCode)
		}
	}

	routes := r.filterRoutesByOriginDestination(origins, destinations)

	if (len(origins) > 0 || len(destinations) > 0) && len(routes) == 0 {
		return nil, fmt.Errorf(
			"destinations %s are not reachable from origins %s",
			strings.Join(destinations, ", "), strings.Join(origins, ", "),
		)
	}

	// Return if no region filtering wanted
	if len(input.DestinationRegions) == 0 {
		return routes, nil
	}

	return r.filterRoutesByRegions(routes, input.DestinationRegions)
}

func (r RouteNetwork) filterRoutesByOriginDestination(origins, destinations []string) []Route {
	// Get ordered wanted origin codes
	var wantedOrigins []string
	if len(origins) == 0 {
		wantedOrigins = slices.Sorted(maps.Keys(r.Routes))
	} else {
		wantedOrigins = origins
		slices.Sort(origins)
	}

	wantedDestinations := mapset.NewSet(destinations...)

	routes := make([]Route, 0)
	for _, wantedOrigin := range wantedOrigins {

		// Get destinations from the origin
		originDestinations := r.Routes[wantedOrigin]
		originDestinationCodes := slices.Sorted(maps.Keys(originDestinations))

		for _, originDestinationCode := range originDestinationCodes {

			// Skip if destination not in wanted destinations
			if len(destinations) > 0 &&
				!wantedDestinations.Contains(originDestinationCode) {
				continue
			}

			routes = append(routes, Route{
				Origin:      r.Airports[wantedOrigin],
				Destination: r.Airports[originDestinationCode],
				Details:     originDestinations[originDestinationCode],
			})
		}
	}

	return routes
}

func (r RouteNetwork) filterRoutesByRegions(routes []Route, wantedRegions []string) ([]Route, error) {
	// Get valid regions
	regions := r.Regions()
	regionsSet := mapset.NewSetWithSize[string](len(regions))
	for _, region := range regions {
		regionsSet.Add(strings.ToLower(region))
	}

	// Validate wanted regions
	wantedRegionSet := mapset.NewSet[string]()
	for _, wantedRegion := range wantedRegions {
		wantedRegion = strings.ToLower(wantedRegion)
		if !regionsSet.Contains(wantedRegion) {
			return nil, fmt.Errorf("invalid region code %q", wantedRegion)
		}

		wantedRegionSet.Add(wantedRegion)
	}

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
