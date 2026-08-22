package avios

import "context"

type Balance struct {
	AvailableAvios int  `json:"availableAvios"`
	IsHousehold    bool `json:"isHousehold"`
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
