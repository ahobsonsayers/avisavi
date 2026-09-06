package gavios

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindFlights_ScanAllDestinations(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/routes",
		httpmock.NewStringResponder(200, routesJSON))

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/allcabins",
		httpmock.NewStringResponder(200, "{}"))

	found, err := client.FindFlights(context.Background(), FindFlightsInput{
		Origins: []string{"lon"},
	})
	require.NoError(t, err)
	require.Len(t, found, 2)

	// Sorted by destination code.
	assert.Equal(t, "ABV", found[0].Destination.AirportCode)
	assert.Equal(t, "JFK", found[1].Destination.AirportCode)
	assert.Equal(t, "LON", found[0].Origin.AirportCode)
}

func TestFindFlights_SingleDestination(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/routes",
		httpmock.NewStringResponder(200, routesJSON))

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/allcabins",
		httpmock.NewStringResponder(200, "{}"))

	found, err := client.FindFlights(context.Background(), FindFlightsInput{
		Origins:      []string{"LON"},
		Destinations: []string{"abv"},
	})
	require.NoError(t, err)
	require.Len(t, found, 1)

	// Airport metadata survives narrowing.
	assert.Equal(t, "London", found[0].Origin.City)
	assert.Equal(t, "Abuja", found[0].Destination.City)
}

func TestFindFlights_UnknownDestination(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/routes",
		httpmock.NewStringResponder(200, routesJSON))

	_, err := client.FindFlights(context.Background(), FindFlightsInput{
		Origins:      []string{"LON"},
		Destinations: []string{"SYD"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown airport code "SYD"`)
}

func TestFindFlights_RegionFilter(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/routes",
		httpmock.NewStringResponder(200, routesJSON))

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/allcabins",
		httpmock.NewStringResponder(200, "{}"))

	found, err := client.FindFlights(context.Background(), FindFlightsInput{
		Origins:            []string{"LON"},
		DestinationRegions: []string{"africa"},
	})
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "ABV", found[0].Destination.AirportCode)
}

func TestFindFlights_UnknownRegion(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/routes",
		httpmock.NewStringResponder(200, routesJSON))

	_, err := client.FindFlights(context.Background(), FindFlightsInput{
		Origins:            []string{"LON"},
		DestinationRegions: []string{"Space"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown region "Space"`)
}

func TestFindFlights_DateFilterFlights(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/routes",
		httpmock.NewStringResponder(200, routesJSON))

	httpmock.RegisterResponder("GET", baseURL+"/spend/v1/flight/allcabins",
		httpmock.NewStringResponder(200, routeFlightsJSON))

	found, err := client.FindFlights(context.Background(), FindFlightsInput{
		Origins:  []string{"LON"},
		Outbound: DateRange{On: testTime(23)},
	})
	require.NoError(t, err)
	require.Len(t, found, 2)
	for _, route := range found {
		economy := route.RouteFlights.Economy
		require.Len(t, economy.Outbound, 2)
		assert.Equal(t, "2026-06-23T08:30:00", economy.Outbound[0].Departure.Format(departureTimeLayout))
		assert.Equal(t, "2026-06-23T21:00:00", economy.Outbound[1].Departure.Format(departureTimeLayout))
	}
}
