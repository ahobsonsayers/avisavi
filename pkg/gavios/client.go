package gavios

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/ahobsonsayers/avisavi/pkg/gavios/auth"
	"github.com/google/uuid"
	"resty.dev/v3"
)

const baseURL = "https://api.rewardsapp.iagl.digital"

type Client struct {
	httpClient *resty.Client
	authData   *auth.AuthData
}

func NewClient(authData *auth.AuthData) *Client {
	restyClient := resty.New()
	restyClient.SetBaseURL(baseURL)
	restyClient.SetHeaders(map[string]string{
		"User-Agent":       "android;V5.35.0 Build:22340160-RN",
		"x-api-programme":  "BAEC",
		"x-api-iso":        "GB",
		"x-auth-type":      "access_token",
		"x-api-version":    "3.2.0",
		"x-auth-client-id": "BAEC-" + uuid.NewString(), // Not validated
		"x-api-key":        "unused",
		"accept":           "application/json",
	})
	restyClient.SetRetryCount(5)
	restyClient.SetRetryWaitTime(100 * time.Millisecond)
	restyClient.SetRetryMaxWaitTime(2 * time.Second)
	restyClient.AddRetryConditions(resty.RetryConditionStatusTooManyRequests)

	return &Client{
		httpClient: restyClient,
		authData:   authData,
	}
}

func (c *Client) MembershipNumber() (string, error) {
	return c.authData.MembershipNumber()
}

func (c *Client) Balance(ctx context.Context) (Balance, error) {
	var balance Balance
	err := c.get(
		ctx,
		"/member/v1/balance",
		nil,
		&balance,
	)
	if err != nil {
		return Balance{}, err
	}

	return balance, nil
}

// RouteNetwork fetches reward destinations grouped by origin airport.
func (c *Client) RouteNetwork(ctx context.Context, adults int, oneWay bool) (RouteNetwork, error) {
	query := url.Values{}
	query.Set("ByAirport", "true")
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("OneWay", strconv.FormatBool(oneWay))

	var routes RouteNetwork
	err := c.get(
		ctx,
		"/spend/v1/flight/routes",
		query,
		&routes,
	)
	if err != nil {
		return RouteNetwork{}, err
	}
	return routes, nil
}

func (c *Client) RouteFlights(
	ctx context.Context,
	origin, destination string,
	oneWay bool,
	adults int,
) (RouteFlights, error) {
	origin, err := NormalizeAirportCode(origin)
	if err != nil {
		return RouteFlights{}, err
	}

	destination, err = NormalizeAirportCode(destination)
	if err != nil {
		return RouteFlights{}, err
	}

	query := url.Values{}
	query.Set("Origin", origin)
	query.Set("Destination", destination)
	query.Set("OneWay", strconv.FormatBool(oneWay))
	query.Set("Adults", strconv.Itoa(adults))
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("IncludeNonBookableFlights", "true")

	var routeFlights RouteFlights
	err = c.get(
		ctx,
		"/spend/v1/flight/allcabins",
		query,
		&routeFlights,
	)
	if err != nil {
		return RouteFlights{}, err
	}

	return routeFlights, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, outValue any) error {
	request := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.authData.AccessToken)

	if len(query) > 0 {
		request.SetQueryParamsFromValues(query)
	}

	response, err := request.Get(path)
	if err != nil {
		return fmt.Errorf("avios request failed: %w", err)
	}

	if !response.IsStatusSuccess() {
		return fmt.Errorf("avios api error: %d: %s", response.StatusCode(), response.String())
	}

	if outValue != nil {
		err = json.Unmarshal(response.Bytes(), outValue)
		if err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}
