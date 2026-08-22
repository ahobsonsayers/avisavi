package avios

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type Availability struct {
	Origin                        LocationCode         `json:"origin"`
	Destination                   LocationCode         `json:"destination"`
	CabinAvailability             map[string]CabinData `json:"cabinAvailability"`
	HighSeatAvailabilityThreshold int                  `json:"highSeatAvailabilityThreshold"`
	MediumAvailabilityThreshold   int                  `json:"mediumAvailabilityThreshold"`
	LowSeatAvailabilityThreshold  int                  `json:"lowSeatAvailabilityThreshold"`
}

type LocationCode struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type CabinData struct {
	Outbound Direction `json:"outbound"`
	Inbound  Direction `json:"inbound"`
}

type Direction struct {
	FromAvios        int                 `json:"fromAvios"`
	AvailabilityFrom string              `json:"availabilityFrom"`
	AvailabilityTo   string              `json:"availabilityTo"`
	AvailableFlights map[string][]Flight `json:"availableFlights"`
}

type Flight struct {
	Date    string `json:"date"`
	Time    string `json:"time"`
	Peak    bool   `json:"peak"`
	Direct  bool   `json:"direct"`
	Avios   int    `json:"avios"`
	Seats   int    `json:"seats"`
	Carrier string `json:"carrier"`
}

func (c *Client) Availability(
	ctx context.Context,
	origin, destination string,
	oneWay bool,
	adults int,
) (Availability, error) {
	query := url.Values{}
	query.Set("Origin", origin)
	query.Set("Destination", destination)
	query.Set("OneWay", strconv.FormatBool(oneWay))
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("IncludeNonBookableFlights", "true")

	membershipNumber, err := c.MembershipNumber()
	if err != nil {
		return Availability{}, err
	}

	var availability Availability
	err = c.get(
		ctx,
		fmt.Sprintf("/spend/v3/programmes/BAEC/GB/%s/flight/availability/allcabins", membershipNumber),
		query, authRaw,
		&availability,
	)
	if err != nil {
		return Availability{}, err
	}

	return availability, nil
}
