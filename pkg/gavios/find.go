package gavios

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

type FindFlightsInput struct {
	// If set finds flights from these origins.
	// If empty finds flights from any origin.
	Origins []string

	// If set finds flights to these destinations.
	// If empty finds flights to any destination.
	Destinations []string

	// If set finds flights to destinations in these regions.
	// If empty finds flights to any region.
	DestinationRegions []string

	Adults int
	OneWay bool

	Outbound DateRange
	Return   DateRange
}

type FoundFlights struct {
	Origin      Airport
	Destination Airport
	Flights     RouteFlights
}

// FindFlights find flights relevant to a filtering input
func (c *Client) FindFlights(ctx context.Context, input FindFlightsInput) ([]FoundFlights, error) {
	// Find routes relevant to the input
	routes, err := c.findRoutes(ctx, input)
	if err != nil {
		return nil, err
	}

	// Find flights for the routes and the input (adults, dater range etc)
	flights := make([]FoundFlights, 0, len(routes))
	for _, route := range routes {
		routeFlights, err := c.RouteFlights(
			ctx,
			route.Origin.AirportCode,
			route.Destination.AirportCode,
			input.OneWay,
			input.Adults,
		)
		if err != nil {
			return nil, err
		}

		flights = append(
			flights,
			FoundFlights{
				Origin:      route.Origin,
				Destination: route.Destination,
				Flights:     routeFlights.FilterByDates(input.Outbound, input.Return),
			},
		)
	}

	return flights, nil
}

// findRoutes find the routes relevant to the find flights input.
// This just finds routes - nothing about tickets or availability.
func (c *Client) findRoutes(ctx context.Context, input FindFlightsInput) ([]Route, error) {
	origins, err := NormalizeAirportCodes(input.Origins)
	if err != nil {
		return nil, err
	}

	destinations, err := NormalizeAirportCodes(input.Destinations)
	if err != nil {
		return nil, err
	}

	network, err := c.RouteNetwork(ctx, input.Adults, input.OneWay)
	if err != nil {
		return nil, err
	}

	// Filter get routes between origins and destinations
	routes := filterByOriginAndDestination(network, origins, destinations)

	if len(routes) == 0 {
		return nil, fmt.Errorf(
			"destinations %s are not reachable from origins %s",
			strings.Join(destinations, ", "), strings.Join(origins, ", "),
		)
	}

	// Return if no region filtering wanted
	if len(input.DestinationRegions) == 0 {
		return routes, nil
	}

	// Get valid regions
	regions := network.Regions()
	regionsSet := mapset.NewSetWithSize[string](len(regions))
	for _, region := range regions {
		regionsSet.Add(strings.ToLower(region))
	}

	// Validate wanted regions
	wantedRegionSet := mapset.NewSet[string]()
	for _, wantedRegion := range input.DestinationRegions {
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

// filterByOriginAndDestination filters the network by origins and destinations, returning
// the routes - or the routes to all origins/destinations when either or neither are set.
// Routes are ordered by origin then destination code.
func filterByOriginAndDestination(network RouteNetwork, origins, destinations []string) []Route {
	// Get ordered wanted origin codes
	var wantedOrigins []string
	if len(origins) == 0 {
		wantedOrigins = slices.Sorted(maps.Keys(network.Routes))
	} else {
		wantedOrigins = origins
		slices.Sort(origins)
	}

	wantedDestinations := mapset.NewSet(destinations...)

	routes := make([]Route, 0)
	for _, wantedOrigin := range wantedOrigins {

		// Get destinations from the origin
		originDestinations := network.Routes[wantedOrigin]
		originDestinationCodes := slices.Sorted(maps.Keys(originDestinations))

		for _, originDestinationCode := range originDestinationCodes {

			// Skip if destination not in wanted destinations
			if len(destinations) > 0 &&
				!wantedDestinations.Contains(originDestinationCode) {
				continue
			}

			routes = append(routes, Route{
				Origin:      network.Airports[wantedOrigin],
				Destination: network.Airports[originDestinationCode],
				Details:     originDestinations[originDestinationCode],
			})
		}
	}

	return routes
}
