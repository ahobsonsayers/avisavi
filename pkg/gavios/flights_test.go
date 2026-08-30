package gavios

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testTime(month time.Month, day int) time.Time {
	return time.Date(2026, month, day, 10, 0, 0, 0, time.UTC)
}

func testRouteFlights() RouteFlights {
	return RouteFlights{
		Economy: TripFlights{
			Outbound: []Flight{
				{Departure: testTime(time.June, 22), Time: "10:00", Seats: 2, Carrier: "BA"},
				{Departure: testTime(time.June, 23), Time: "22:30", Seats: 9, Carrier: "BA"},
				{Departure: testTime(time.June, 24), Time: "07:15", Seats: 4, Carrier: "BA"},
			},
			Inbound: []Flight{
				{Departure: testTime(time.June, 25), Time: "18:00", Seats: 3, Carrier: "BA"},
			},
		},
	}
}

func TestDateRangeIsZero(t *testing.T) {
	assert.True(t, DateRange{}.IsZero())
	assert.False(t, DateRange{On: testTime(time.September, 9)}.IsZero())
}

func TestDateRangeInRange(t *testing.T) {
	day := testTime(time.June, 23)

	assert.True(t, DateRange{On: day}.InRange(day))
	assert.False(t, DateRange{On: testTime(time.June, 24)}.InRange(day))

	// After and Before are exclusive bounds.
	assert.True(t, DateRange{After: testTime(time.June, 22)}.InRange(day))
	assert.False(t, DateRange{After: day}.InRange(day))
	assert.True(t, DateRange{Before: testTime(time.June, 24)}.InRange(day))
	assert.False(t, DateRange{Before: day}.InRange(day))
	assert.False(t, DateRange{After: testTime(time.June, 22), Before: day}.InRange(day))

	// Zero range matches everything.
	assert.True(t, DateRange{}.InRange(day))
}

func TestFlightUnmarshalJSON(t *testing.T) {
	flightJSON := `{"date":"2026-06-22T10:00:00","time":"10:00","seats":2,"carrier":"BA"}`

	var flight Flight
	err := json.Unmarshal([]byte(flightJSON), &flight)
	require.NoError(t, err)

	assert.Equal(t, testTime(time.June, 22), flight.Departure)
	assert.Equal(t, "10:00", flight.Time)
	assert.Equal(t, 2, flight.Seats)
	assert.Equal(t, "BA", flight.Carrier)
}

func TestRouteFlightsFilterByDates(t *testing.T) {
	routeFlights := testRouteFlights()

	filtered := routeFlights.FilterByDates(
		DateRange{On: testTime(time.June, 23)},
		DateRange{On: testTime(time.June, 25)},
	)
	require.Len(t, filtered.Economy.Outbound, 1)
	assert.Equal(t, "22:30", filtered.Economy.Outbound[0].Time)
	require.Len(t, filtered.Economy.Inbound, 1)
	assert.Equal(t, "18:00", filtered.Economy.Inbound[0].Time)

	unchanged := routeFlights.FilterByDates(DateRange{}, DateRange{})
	assert.Equal(t, routeFlights, unchanged)
}