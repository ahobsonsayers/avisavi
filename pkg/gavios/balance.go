package gavios

import "context"

// Balance holds the current Avios balance for the account.
type Balance struct {
	// AvailableAvios is the number of Avios currently available to spend.
	AvailableAvios int `json:"availableAvios"`
	// IsHousehold reports whether the balance is a household (shared) balance.
	IsHousehold bool `json:"isHousehold"`
}

func (c *Client) Balance(ctx context.Context) (Balance, error) {
	var balance Balance
	err := c.get(
		ctx,
		"/member/v1/balance",
		nil,
		authBearer,
		&balance,
	)
	if err != nil {
		return Balance{}, err
	}

	return balance, nil
}
