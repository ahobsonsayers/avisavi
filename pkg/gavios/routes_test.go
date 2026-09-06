package gavios

import (
	"encoding/json"
	"testing"
)

const routesJSON = `{"origins":[` +
	`{"airportCode":"LON","airportName":"London Heathrow","name":"London","countryName":"United Kingdom",` +
	`"destinations":[` +
	`{"airportCode":"ABV","airportName":"Nnamdi Azikiwe","name":"Abuja",` +
	`"countryCode":"NG","countryName":"Nigeria",` +
	`"broadSearchCategories":["Africa"],` +
	`"aviosPerCabinClass":{"Economy":{"min":100,"max":200}}},` +
	`{"airportCode":"JFK","airportName":"John F Kennedy","name":"New York",` +
	`"countryCode":"US","countryName":"United States",` +
	`"broadSearchCategories":["North America"],` +
	`"aviosPerCabinClass":{"Business":{"min":50000,"max":70000}}}` +
	`]},` +
	`{"airportCode":"MAN","airportName":"Manchester","name":"Manchester","countryName":"United Kingdom",` +
	`"destinations":[` +
	`{"airportCode":"DUB","airportName":"Dublin","name":"Dublin",` +
	`"countryCode":"IE","countryName":"Ireland",` +
	`"aviosPerCabinClass":{"Economy":{"min":7500,"max":12500}}}` +
	`]}]}`

func unmarshalTestRouteNetwork(t *testing.T) RouteNetwork {
	t.Helper()

	var response routesResponse
	err := json.Unmarshal([]byte(routesJSON), &response)
	if err != nil {
		t.Fatalf("unmarshal routes: %v", err)
	}

	return response.toRouteNetwork()
}

func TestRoutes_UnmarshalJSON_Airports(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	if len(routes.Airports) != 5 {
		t.Fatalf("expected 5 airports, got %d", len(routes.Airports))
	}

	origin := routes.Airports["LON"]
	if origin.City != "London" || origin.CountryCode != "" || origin.Country != "United Kingdom" {
		t.Errorf("origin LON metadata wrong: %+v", origin)
	}

	destination := routes.Airports["ABV"]
	if destination.CountryCode != "NG" || destination.Country != "Nigeria" {
		t.Errorf("destination ABV metadata wrong: %+v", destination)
	}
}

func TestRoutes_UnmarshalJSON_Routes(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	abvRoute, found := routes.Routes["LON"]["ABV"]
	if !found {
		t.Fatal("route LON->ABV missing")
	}
	if abvRoute.AviosPrices.Economy.MaxAvios != 200 || abvRoute.AviosPrices.Business.MinAvios != 0 {
		t.Errorf("LON->ABV cabin prices wrong: %+v", abvRoute)
	}
	if abvRoute.Region != "Africa" {
		t.Errorf("LON->ABV region wrong: %q", abvRoute.Region)
	}

	jfkRoute := routes.Routes["LON"]["JFK"]
	if jfkRoute.AviosPrices.Business.MaxAvios != 70000 || jfkRoute.AviosPrices.Economy.MinAvios != 0 {
		t.Errorf("LON->JFK cabin prices wrong: %+v", jfkRoute)
	}
}

func TestRoutes_Regions(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	regions := routes.Regions()
	if len(regions) != 2 || regions[0] != "Africa" || regions[1] != "North America" {
		t.Errorf("regions wrong: %+v", regions)
	}

	emptyRegions := RouteNetwork{}.Regions()
	if emptyRegions != nil {
		t.Error("empty routes should have no regions")
	}
}

func TestRoutes_UnmarshalJSON_Empty(t *testing.T) {
	var response routesResponse
	err := json.Unmarshal([]byte(`{"origins":[]}`), &response)
	if err != nil {
		t.Fatalf("unmarshal routes: %v", err)
	}

	routes := response.toRouteNetwork()

	if routes.Airports == nil || routes.Routes == nil {
		t.Error("maps should be initialised, not nil")
	}
}

func TestRoutes_FindRoutes_All(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	allRoutes, err := routes.FindRoutes(FindRoutesInput{})
	if err != nil {
		t.Fatalf("FindRoutes(FindRoutesInput{}): %v", err)
	}
	if len(allRoutes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(allRoutes))
	}
	// Sorted by origin, then destination.
	wantOrder := [][2]string{{"LON", "ABV"}, {"LON", "JFK"}, {"MAN", "DUB"}}
	for i, want := range wantOrder {
		got := allRoutes[i]
		if got.Origin.AirportCode != want[0] || got.Destination.AirportCode != want[1] {
			t.Errorf("route %d wrong: got %s->%s, want %s->%s",
				i, got.Origin.AirportCode, got.Destination.AirportCode, want[0], want[1])
		}
	}
	if allRoutes[0].Destination.City != "Abuja" {
		t.Errorf("route airport metadata missing: %+v", allRoutes[0].Destination)
	}
}

func TestRoutes_FindRoutes_Origin(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	lonRoutes, err := routes.FindRoutes(FindRoutesInput{Origins: []string{"lon"}})
	if err != nil {
		t.Fatalf("FindRoutes(lon): %v", err)
	}
	if len(lonRoutes) != 2 {
		t.Fatalf("expected 2 LON routes, got %d", len(lonRoutes))
	}
	if lonRoutes[0].Origin.AirportCode != "LON" {
		t.Errorf("origin missing from route: %+v", lonRoutes[0].Origin)
	}
	if lonRoutes[0].Details.Region != "Africa" {
		t.Errorf("route details missing: %+v", lonRoutes[0].Details)
	}

	_, err = routes.FindRoutes(FindRoutesInput{Origins: []string{"JFK"}})
	if err == nil {
		t.Error("unknown origin should error")
	}
	_, err = routes.FindRoutes(FindRoutesInput{Origins: []string{"12"}})
	if err == nil {
		t.Error("invalid code should error")
	}
}

func TestRoutes_FindRoutes_Destination(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	dubRoutes, err := routes.FindRoutes(FindRoutesInput{Destinations: []string{"dub"}})
	if err != nil {
		t.Fatalf("FindRoutes(dub): %v", err)
	}
	if len(dubRoutes) != 1 {
		t.Fatalf("expected 1 DUB route, got %d", len(dubRoutes))
	}
	if dubRoutes[0].Origin.AirportCode != "MAN" || dubRoutes[0].Destination.AirportCode != "DUB" {
		t.Errorf("airports wrong: %+v -> %+v", dubRoutes[0].Origin, dubRoutes[0].Destination)
	}

	_, err = routes.FindRoutes(FindRoutesInput{Destinations: []string{"SYD"}})
	if err == nil {
		t.Error("unreachable destination should error")
	}
}

func TestRoutes_FindRoutes_Region(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	africaRoutes, err := routes.FindRoutes(FindRoutesInput{DestinationRegions: []string{"africa"}})
	if err != nil {
		t.Fatalf("FindRoutes(africa): %v", err)
	}
	if len(africaRoutes) != 1 {
		t.Fatalf("expected 1 Africa route, got %d", len(africaRoutes))
	}
	if africaRoutes[0].Destination.AirportCode != "ABV" {
		t.Errorf("wrong route: %+v", africaRoutes[0])
	}

	_, err = routes.FindRoutes(FindRoutesInput{DestinationRegions: []string{"Oceania"}})
	if err == nil {
		t.Error("invalid region should error")
	}
}
