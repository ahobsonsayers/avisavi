package gavios

import (
	"context"
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
	Origin       Airport
	Destination  Airport
	RouteFlights RouteFlights
}

// FindFlights find flights relevant to a filtering input
func (c *Client) FindFlights(ctx context.Context, input FindFlightsInput) ([]FoundFlights, error) {
	// Find routes relevant to the input
	network, err := c.Network(ctx)
	if err != nil {
		return nil, err
	}

	routes, err := network.FindRoutes(
		FindRoutesInput{
			Origins:            input.Origins,
			Destinations:       input.Destinations,
			DestinationRegions: input.DestinationRegions,
		},
	)
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

		flightInDates := routeFlights.FilterByDates(input.Outbound, input.Return)

		flights = append(
			flights,
			FoundFlights{
				Origin:       route.Origin,
				Destination:  route.Destination,
				RouteFlights: flightInDates,
			},
		)
	}

	return flights, nil
}
