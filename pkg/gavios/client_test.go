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

const routesResp = `{"origins":[{"originAirportCode":"LON","destinations":[` +
	`{"airportCode":"ABV","name":"Abuja","countryName":"Nigeria",` +
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

	routes, err := client.Routes(context.Background(), "LON", 1, false)
	require.NoError(t, err)
	require.Len(t, routes.Origins, 1)
	assert.Equal(t, "LON", routes.Origins[0].OriginAirportCode)
	require.Len(t, routes.Origins[0].Destinations, 1)
	destination := routes.Origins[0].Destinations[0]
	assert.Equal(t, "ABV", destination.DestinationAirportCode)
	assert.Equal(t, "Abuja", destination.DestinationName)
	assert.Equal(t, 100, destination.AviosPerCabinClass["Economy"].Min)

	require.NotNil(t, captured)
	query := captured.URL.Query()
	assert.Equal(t, "LON", query.Get("OriginAirport"))
	assert.Equal(t, "true", query.Get("ByAirport"))
	assert.Equal(t, "1", query.Get("Adults"))
	assert.Equal(t, "false", query.Get("OneWay"))
}

const availabilityJSON = `{"availabilityPerCabin":{"Economy":{"outbound":{"flightsPerDate":{"2026-06-23T00:00:00":[{"date":"2026-06-23T22:30:00","time":"22:30","seats":9,"carrier":"BA"}]}},"inbound":{"flightsPerDate":{}}}}}`

func TestClient_Availability(t *testing.T) {
	client := testClient()
	defer httpmock.DeactivateAndReset()

	path := "https://api.rewardsapp.iagl.digital/spend/v1/flight/allcabins"
	var captured *http.Request
	httpmock.RegisterResponder("GET", path,
		func(req *http.Request) (*http.Response, error) {
			captured = req
			return httpmock.NewStringResponse(200, availabilityJSON), nil
		})

	availability, err := client.Availability(context.Background(), "LON", "ABV", false, 1)
	require.NoError(t, err)
	economy, ok := availability.CabinAvailability["Economy"]
	require.True(t, ok)
	require.Contains(t, economy.Outbound.Flights, "2026-06-23T00:00:00")
	flight := economy.Outbound.Flights["2026-06-23T00:00:00"][0]
	assert.Equal(t, 9, flight.Seats)

	require.NotNil(t, captured)
	query := captured.URL.Query()
	assert.Equal(t, "LON", query.Get("Origin"))
	assert.Equal(t, "ABV", query.Get("Destination"))
	assert.Equal(t, "false", query.Get("OneWay"))
	assert.Equal(t, "1", query.Get("Adults"))
	assert.Equal(t, "true", query.Get("IncludeNonBookableFlights"))
}

func TestNormalizeAirportCode(t *testing.T) {
	upper, err := normalizeAirportCode("lon")
	require.NoError(t, err)
	assert.Equal(t, "LON", upper)

	for _, bad := range []string{"", "LO", "LONN", "L0N", "12A"} {
		_, err := normalizeAirportCode(bad)
		assert.Error(t, err, "code %q should be rejected", bad)
	}
}
