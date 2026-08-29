package gavios

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoutes_Regions(t *testing.T) {
	routes := Routes{
		Origins: []Origin{
			{
				AirportCode: "LON",
				Destinations: []Destination{
					{DestinationAirportCode: "ABV", BroadSearchCategories: []string{"Beach", "City break"}},
					{DestinationAirportCode: "JFK", BroadSearchCategories: []string{"City break"}},
				},
			},
			{
				AirportCode: "DUB",
				Destinations: []Destination{
					{DestinationAirportCode: "NBO", BroadSearchCategories: []string{"Safari"}},
				},
			},
		},
	}

	assert.Equal(t, []string{"Beach", "City break", "Safari"}, routes.Regions())

	assert.Empty(t, Routes{}.Regions())
}
