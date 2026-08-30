package gavios

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/ahobsonsayers/avisavi/pkg/gavios/auth"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JWT whose payload decodes to membership_id "01234567".
const testToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwczovL2F2aW9zLmNvbS9tZW1iZXJzaGlwX2lkIjoiMDEyMzQ1NjcifQ.ZmFrZXNpZ25hdHVyZQ"

func testClient() *Client {
	c := NewClient(&auth.AuthData{AccessToken: testToken})
	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	c.httpClient.SetTransport(hc.Transport)
	return c
}

func TestClient_AuthHeader(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"balance", "/member/v1/balance"},
		{"routes", "/spend/v1/flight/routes"},
		{"allcabins", "/spend/v1/flight/allcabins"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testClient()
			defer httpmock.DeactivateAndReset()

			var captured *http.Request
			httpmock.RegisterResponder("GET", baseURL+tt.path,
				func(req *http.Request) (*http.Response, error) {
					captured = req
					return httpmock.NewStringResponse(200, `{}`), nil
				})

			var out map[string]any
			require.NoError(t, client.get(context.Background(), tt.path, nil, &out))

			require.NotNil(t, captured)
			assert.Equal(t, "Bearer "+testToken, captured.Header.Get("Authorization"))
			assert.Equal(t, "BAEC", captured.Header.Get("x-api-programme"))
			assert.True(t, strings.HasPrefix(captured.Header.Get("x-auth-client-id"), "BAEC-"))
			assert.Equal(t, "unused", captured.Header.Get("x-api-key"))
		})
	}
}

const routesResp = `{"origins":[{"airportCode":"LON","airportName":"London Heathrow","name":"London","countryName":"United Kingdom",` +
	`"destinations":[` +
	`{"airportCode":"ABV","airportName":"Nnamdi Azikiwe International","name":"Abuja","countryCode":"NG","countryName":"Nigeria",` +
	`"aviosPerCabinClass":{"Economy":{"min":100,"max":200}}}]}],"broadSearchGroups":[]}`

func TestClient_Routes(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	var captured *http.Request
	httpmock.RegisterResponder("GET",
		"https://api.rewardsapp.iagl.digital/spend/v1/flight/routes",
		func(req *http.Request) (*http.Response, error) {
			captured = req
			return httpmock.NewStringResponse(200, routesResp), nil
		})

	routes, err := client.RouteNetwork(context.Background(), 1, false)
	require.NoError(t, err)
	require.Len(t, routes.Airports, 2)
	assert.Equal(t, Airport{AirportCode: "LON", AirportName: "London Heathrow", City: "London", Country: "United Kingdom"}, routes.Airports["LON"])
	require.Contains(t, routes.Routes, "LON")
	destination, found := routes.Routes["LON"]["ABV"]
	require.True(t, found, "route LON->ABV should exist")
	assert.Equal(t, "Abuja", routes.Airports["ABV"].City)
	assert.Equal(t, 100, destination.AviosPrices.Economy.MinAvios)

	require.NotNil(t, captured)
	query := captured.URL.Query()
	assert.Equal(t, "true", query.Get("ByAirport"))
	assert.Equal(t, "1", query.Get("Adults"))
	assert.Equal(t, "false", query.Get("OneWay"))

	upper, err := NormalizeAirportCode("lon")
	require.NoError(t, err)
	assert.Equal(t, "LON", upper)
}

const routeFlightsJSON = `{"availabilityPerCabin":{"Economy":{"outbound":{"flightsPerDate":` +
	`{"2026-06-24T00:00:00":[{"date":"2026-06-24T07:15:00","time":"07:15","seats":4,"carrier":"BA"}],` +
	`"2026-06-23T00:00:00":[{"date":"2026-06-23T21:00:00","time":"21:00","seats":2,"carrier":"BA"},` +
	`{"date":"2026-06-23T08:30:00","time":"08:30","seats":9,"carrier":"BA"}]}}` +
	`,"inbound":{"flightsPerDate":{}}}}}`

func TestClient_RouteFlights(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	var captured *http.Request
	httpmock.RegisterResponder("GET",
		"https://api.rewardsapp.iagl.digital/spend/v1/flight/allcabins",
		func(req *http.Request) (*http.Response, error) {
			captured = req
			return httpmock.NewStringResponse(200, routeFlightsJSON), nil
		})

	routeFlights, err := client.RouteFlights(context.Background(), "LON", "ABV", false, 1)
	require.NoError(t, err)

	// Flights are ordered by full departure timestamp.
	economy := routeFlights.Economy
	require.Len(t, economy.Outbound, 3)
	assert.Equal(t, "2026-06-23T08:30:00", economy.Outbound[0].Departure.Format(departureTimeLayout))
	assert.Equal(t, "2026-06-23T21:00:00", economy.Outbound[1].Departure.Format(departureTimeLayout))
	assert.Equal(t, "2026-06-24T07:15:00", economy.Outbound[2].Departure.Format(departureTimeLayout))

	assert.Empty(t, routeFlights.Business.Outbound)

	require.NotNil(t, captured)
	query := captured.URL.Query()
	assert.Equal(t, "LON", query.Get("Origin"))
	assert.Equal(t, "ABV", query.Get("Destination"))
	assert.Equal(t, "false", query.Get("OneWay"))
	assert.Equal(t, "1", query.Get("Adults"))
	assert.Equal(t, "true", query.Get("IncludeNonBookableFlights"))
}

func TestNormalizeAirportCode(t *testing.T) {
	upper, err := NormalizeAirportCode("lon")
	require.NoError(t, err)
	assert.Equal(t, "LON", upper)

	for _, bad := range []string{"", "LO", "LONN", "L0N", "12A"} {
		_, err := NormalizeAirportCode(bad)
		assert.Error(t, err, "code %q should be rejected", bad)
	}
}

func TestClient_RetryOn429(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	calls := 0
	httpmock.RegisterResponder("GET",
		"https://api.rewardsapp.iagl.digital/spend/v1/flight/allcabins",
		func(req *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return httpmock.NewStringResponse(429, `{"message":"too many requests"}`), nil
			}
			return httpmock.NewStringResponse(200, routeFlightsJSON), nil
		})

	_, err := client.RouteFlights(context.Background(), "LON", "ABV", false, 1)
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}
