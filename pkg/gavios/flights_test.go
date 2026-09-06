package gavios

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTime(day int) time.Time {
	return time.Date(2026, time.June, day, 10, 0, 0, 0, time.UTC)
}

func testRouteFlights() RouteFlights {
	return RouteFlights{
		Economy: TripFlights{
			Outbound: []Flight{
				{Departure: testTime(22), Time: "10:00", Seats: 2, Carrier: "BA"},
				{Departure: testTime(23), Time: "22:30", Seats: 9, Carrier: "BA"},
				{Departure: testTime(24), Time: "07:15", Seats: 4, Carrier: "BA"},
			},
			Inbound: []Flight{
				{Departure: testTime(25), Time: "18:00", Seats: 3, Carrier: "BA"},
			},
		},
	}
}

func TestDateRangeInRange(t *testing.T) {
	day := testTime(23)

	assert.True(t, DateRange{On: day}.InRange(day))
	assert.False(t, DateRange{On: testTime(24)}.InRange(day))

	// After and Before are exclusive bounds.
	assert.True(t, DateRange{After: testTime(22)}.InRange(day))
	assert.False(t, DateRange{After: day}.InRange(day))
	assert.True(t, DateRange{Before: testTime(24)}.InRange(day))
	assert.False(t, DateRange{Before: day}.InRange(day))
	assert.False(t, DateRange{After: testTime(22), Before: day}.InRange(day))

	// Zero range matches everything.
	assert.True(t, DateRange{}.InRange(day))
}

func TestNewFlights(t *testing.T) {
	response := flightsPerDateResponse{
		Flights: map[string][]flightResponse{
			"2026-06-22T00:00:00": {
				{Date: "2026-06-22T10:00:00", Time: "10:00", Seats: 2, Carrier: "BA"},
			},
		},
	}

	flights, err := response.toFlights()
	require.NoError(t, err)

	require.Len(t, flights, 1)
	assert.Equal(t, testTime(22), flights[0].Departure)

	bad := flightsPerDateResponse{
		Flights: map[string][]flightResponse{
			"x": {{Date: "not-a-date"}},
		},
	}
	_, err = bad.toFlights()
	assert.Error(t, err)
}

func TestRouteFlightsFilterByDates(t *testing.T) {
	routeFlights := testRouteFlights()

	filtered := routeFlights.FilterByDates(
		DateRange{On: testTime(23)},
		DateRange{On: testTime(25)},
	)

	economy := filtered.Economy
	require.Len(t, economy.Outbound, 1)
	assert.Equal(t, "22:30", economy.Outbound[0].Time)
	require.Len(t, economy.Inbound, 1)
	assert.Equal(t, "18:00", economy.Inbound[0].Time)

	unchanged := routeFlights.FilterByDates(DateRange{}, DateRange{})
	assert.Equal(t, routeFlights, unchanged)
}
