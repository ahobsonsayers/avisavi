package gavios

// Balance holds the current Avios balance for the account.
type Balance struct {
	// AvailableAvios is the number of Avios currently available to spend.
	AvailableAvios int `json:"availableAvios"`
	// IsHousehold reports whether the balance is a household (shared) balance.
	IsHousehold bool `json:"isHousehold"`
}
