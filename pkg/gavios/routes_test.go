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
	`]}]}`

func unmarshalTestRoutes(t *testing.T) Routes {
	t.Helper()

	var routes Routes
	err := json.Unmarshal([]byte(routesJSON), &routes)
	if err != nil {
		t.Fatalf("unmarshal routes: %v", err)
	}

	return routes
}

func TestRoutes_UnmarshalJSON_Airports(t *testing.T) {
	routes := unmarshalTestRoutes(t)

	if len(routes.Airports) != 3 {
		t.Fatalf("expected 3 airports, got %d", len(routes.Airports))
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
	routes := unmarshalTestRoutes(t)

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
	routes := unmarshalTestRoutes(t)

	regions := routes.Regions()
	if len(regions) != 2 || regions[0] != "Africa" || regions[1] != "North America" {
		t.Errorf("regions wrong: %+v", regions)
	}

	emptyRegions := Routes{}.Regions()
	if emptyRegions != nil {
		t.Error("empty routes should have no regions")
	}
}

func TestRoutes_UnmarshalJSON_Empty(t *testing.T) {
	var routes Routes
	err := json.Unmarshal([]byte(`{"origins":[]}`), &routes)
	if err != nil {
		t.Fatalf("unmarshal routes: %v", err)
	}

	if routes.Airports == nil || routes.Routes == nil {
		t.Error("maps should be initialised, not nil")
	}
}
