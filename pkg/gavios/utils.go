package gavios

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

var airportCodeRegex = regexp.MustCompile(`^[a-zA-Z]{3}$`)

// NormalizeAirportCode validates an IATA airport code (3 letters) and
// returns it uppercased. It returns an error for anything else.
func NormalizeAirportCode(code string) (string, error) {
	if !airportCodeRegex.MatchString(code) {
		return "", fmt.Errorf("invalid airport code %q: must be 3 letters", code)
	}
	return strings.ToUpper(code), nil
}

// NormalizeAirportCodes is the same as NormalizeAirportCode for a slice
func NormalizeAirportCodes(codes []string) ([]string, error) {
	normalizedCodes := make([]string, 0, len(codes))
	for _, code := range codes {
		normalizedCode, err := NormalizeAirportCode(code)
		if err != nil {
			return nil, err
		}
		normalizedCodes = append(normalizedCodes, normalizedCode)
	}
	return lo.Uniq(normalizedCodes), nil
}
