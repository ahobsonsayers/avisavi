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

	var routes RouteNetwork
	err := json.Unmarshal([]byte(routesJSON), &routes)
	if err != nil {
		t.Fatalf("unmarshal routes: %v", err)
	}

	return routes
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
	var routes RouteNetwork
	err := json.Unmarshal([]byte(`{"origins":[]}`), &routes)
	if err != nil {
		t.Fatalf("unmarshal routes: %v", err)
	}

	if routes.Airports == nil || routes.Routes == nil {
		t.Error("maps should be initialised, not nil")
	}
}

func TestRoutes_GetRoute(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	route, err := routes.GetRoute("lon", "abv")
	if err != nil {
		t.Fatalf("GetRoute(lon, abv): %v", err)
	}

	if route.Origin.AirportCode != "LON" || route.Destination.AirportCode != "ABV" {
		t.Errorf("airports wrong: origin=%+v destination=%+v", route.Origin, route.Destination)
	}
	if route.Details.Region != "Africa" {
		t.Errorf("details wrong: %+v", route.Details)
	}

	if _, err := routes.GetRoute("JFK", "ABV"); err == nil {
		t.Error("unknown origin should error")
	}
	if _, err := routes.GetRoute("LON", "LON"); err == nil {
		t.Error("unknown destination should error")
	}
	if _, err := routes.GetRoute("XYZ", "ABV"); err == nil {
		t.Error("invalid code should error")
	}
}

func TestRoutes_GetRoutes_All(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	allRoutes, err := routes.GetRoutes("")
	if err != nil {
		t.Fatalf("GetRoutes(): %v", err)
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

func TestRoutes_GetRoutes_Origin(t *testing.T) {
	routes := unmarshalTestRouteNetwork(t)

	lonRoutes, err := routes.GetRoutes("lon")
	if err != nil {
		t.Fatalf("GetRoutes(lon): %v", err)
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

	if _, err := routes.GetRoutes("JFK"); err == nil {
		t.Error("unknown origin should error")
	}
	if _, err := routes.GetRoutes("12"); err == nil {
		t.Error("invalid code should error")
	}
}
