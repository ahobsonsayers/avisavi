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
	data, err := c.get(
		ctx,
		"/member/v1/balance",
		nil,
	)
	if err != nil {
		return Balance{}, err
	}

	var balance Balance
	err = json.Unmarshal(data, &balance)
	if err != nil {
		return Balance{}, fmt.Errorf("decoding balance response: %w", err)
	}

	return balance, nil
}

func (c *Client) RouteNetwork(ctx context.Context) (RouteNetwork, error) {
	query := url.Values{}
	query.Set("ByAirport", "true")
	query.Set("Adults", "1")
	query.Set("YoungAdults", "0")
	query.Set("Children", "0")
	query.Set("Infants", "0")
	query.Set("OneWay", "false")

	data, err := c.get(
		ctx,
		"/spend/v1/flight/routes",
		query,
	)
	if err != nil {
		return RouteNetwork{}, err
	}

	var response routesResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return RouteNetwork{}, fmt.Errorf("decoding routes response: %w", err)
	}

	return response.toRouteNetwork(), nil
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
	query.Set("IncludeNonBookableFlights", "false")

	data, err := c.get(
		ctx,
		"/spend/v1/flight/allcabins",
		query,
	)
	if err != nil {
		return RouteFlights{}, err
	}

	var response routeFlightsResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return RouteFlights{}, fmt.Errorf("decoding flights response: %w", err)
	}

	return response.toRouteFlights()
}

func (c *Client) get(ctx context.Context, path string, query url.Values) (json.RawMessage, error) {
	request := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+c.authData.AccessToken)

	if len(query) > 0 {
		request.SetQueryParamsFromValues(query)
	}

	response, err := request.Get(path)
	if err != nil {
		return nil, fmt.Errorf("avios request failed: %w", err)
	}

	if !response.IsStatusSuccess() {
		return nil, fmt.Errorf("avios api error: %d: %s", response.StatusCode(), response.String())
	}

	return response.Bytes(), nil
}
