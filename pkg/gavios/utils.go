package gavios

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/samber/lo"
)

var airportCodeRegex = regexp.MustCompile(`^[a-zA-Z]{3}$`)

// NormalizeAirportCode validates an IATA airport code (3 letters) and returns it uppercased.
// If validCodes is set,  also validates the code is in the valid codes
func NormalizeAirportCode(code string, validCodes []string) (string, error) {
	if !airportCodeRegex.MatchString(code) {
		return "", fmt.Errorf("invalid airport code %q: must be 3 letters", code)
	}

	normalized := strings.ToUpper(code)
	if len(validCodes) > 0 && !slices.Contains(validCodes, normalized) {
		return "", fmt.Errorf("unknown airport code %q", code)
	}

	return normalized, nil
}

// NormalizeAirportCodes is the same as NormalizeAirportCode for a slice
func NormalizeAirportCodes(codes, validCodes []string) ([]string, error) {
	normalizedCodes := make([]string, 0, len(codes))
	for _, code := range codes {
		normalizedCode, err := NormalizeAirportCode(code, validCodes)
		if err != nil {
			return nil, err
		}

		normalizedCodes = append(normalizedCodes, normalizedCode)
	}

	return lo.Uniq(normalizedCodes), nil
}

// NormalizeRegion lowercases a region.
// If validRegions is set, also validates the region is in the valid regions
func NormalizeRegion(region string, validRegions []string) (string, error) {
	normalizedRegion := strings.ToLower(region)

	if len(validRegions) == 0 {
		return normalizedRegion, nil
	}

	for _, validRegion := range validRegions {
		if strings.EqualFold(validRegion, normalizedRegion) {
			return normalizedRegion, nil
		}
	}

	return "", fmt.Errorf("unknown region %q", region)
}

// NormalizeRegions is the same as NormalizeRegion for a slice
func NormalizeRegions(regions, validRegions []string) ([]string, error) {
	normalizedRegions := make([]string, 0, len(regions))
	for _, region := range regions {
		normalizedRegion, err := NormalizeRegion(region, validRegions)
		if err != nil {
			return nil, err
		}
		normalizedRegions = append(normalizedRegions, normalizedRegion)
	}
	return lo.Uniq(normalizedRegions), nil
}
