package avios

import (
	"context"
	"net/url"
	"strconv"
)

type Routes struct {
	Origins               []Origin   `json:"origins"`
	BroadSearchCategories []Category `json:"broadSearchCategories"`
}

type Origin struct {
	OriginAirportCode string        `json:"originAirportCode"`
	Destinations      []Destination `json:"destinations"`
}

type Destination struct {
	DestinationAirportCode string                `json:"destinationAirportCode"`
	DestinationAirportName string                `json:"destinationAirportName"`
	DestinationName        string                `json:"destinationName"`
	CountryCode            string                `json:"countryCode"`
	CountryName            string                `json:"countryName"`
	BroadSearchCategories  []string              `json:"broadSearchCategories"`
	AviosPerCabinClass     map[string]CabinPrice `json:"aviosPerCabinClass"`
	FlownByPartners        []string              `json:"flownByPartners"`
}

type CabinPrice struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type Category struct {
	Name               string                `json:"name"`
	DestinationCount   int                   `json:"destinationCount"`
	AviosPerCabinClass map[string]CabinPrice `json:"aviosPerCabinClass"`
}

func (c *Client) Routes(ctx context.Context, originAirport string, adults int, oneWay bool) (Routes, error) {
	query := url.Values{}
	query.Set("ByAirport", "true")
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("OneWay", strconv.FormatBool(oneWay))
	if originAirport != "" {
		query.Set("OriginAirport", originAirport)
	}

	var routes Routes
	err := c.get(
		ctx,
		"/spend/v3/programmes/BAEC/GB/flight/routes",
		query,
		authRaw,
		&routes,
	)
	if err != nil {
		return Routes{}, err
	}
	return routes, nil
}
