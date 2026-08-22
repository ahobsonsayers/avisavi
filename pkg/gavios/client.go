package gavios

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ahobsonsayers/avios-cli/pkg/gavios/auth"
	"github.com/google/uuid"
	"resty.dev/v3"
)

const baseURL = "https://api.rewardsapp.iagl.digital"

type authMode string

const (
	authRaw    authMode = ""        // spend/v3, whitelabel/v3: raw JWT, no Bearer prefix
	authBearer authMode = "Bearer " // member/v1, collect/v1, alerts/v1, etc.
)

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

	return &Client{
		httpClient: restyClient,
		authData:   authData,
	}
}

func (c *Client) MembershipNumber() (string, error) {
	return c.authData.MembershipNumber()
}

func (c *Client) get(ctx context.Context, path string, query url.Values, authMode authMode, outValue any) error {
	request := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", string(authMode)+c.authData.AccessToken)

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

var airportCodeRegex = regexp.MustCompile(`^[a-zA-Z]{3}$`)

// normalizeAirportCode validates an IATA airport code (3 letters) and
// returns it uppercased. It returns an error for anything else.
func normalizeAirportCode(code string) (string, error) {
	if !airportCodeRegex.MatchString(code) {
		return "", fmt.Errorf("invalid airport code %q: must be 3 letters", code)
	}
	return strings.ToUpper(code), nil
}
