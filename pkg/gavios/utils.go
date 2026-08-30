package gavios

import (
	"fmt"
	"regexp"
	"strings"
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
