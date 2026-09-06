package main

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"charm.land/lipgloss/v2/table"
)

// Shared command styles.

var (
	successStyle = lipgloss.NewStyle().Foreground(adaptiveColour("40"))
	errorStyle   = lipgloss.NewStyle().Foreground(adaptiveColour("196"))
	arrowStyle   = lipgloss.NewStyle().Foreground(adaptiveColour("248"))
	numberStyle  = lipgloss.NewStyle().Bold(true)
	unitStyle    = lipgloss.NewStyle().Bold(true).Foreground(adaptiveColour("248"))
	noteStyle    = lipgloss.NewStyle().Foreground(adaptiveColour("248"))
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(adaptiveColour("248"))

	// Airport codes in routes output: origin cyan, destination green.
	originCodeStyle      = lipgloss.NewStyle().Foreground(adaptiveColour("80"))
	destinationCodeStyle = lipgloss.NewStyle().Foreground(adaptiveColour("114"))
)

// Find command styles.

var (
	directionStyle = lipgloss.NewStyle().Foreground(adaptiveColour("248"))
	carrierStyle   = lipgloss.NewStyle().Faint(true)

	cabinStyles = map[string]lipgloss.Style{
		"Economy":  cabinStyle("136"),
		"Premium":  cabinStyle("170"),
		"Business": cabinStyle("62"),
		"First":    cabinStyle("205"),
	}

	seatGreen  = seatColour("46")
	seatYellow = seatColour("226")
	seatRed    = seatColour("196")
)

// baseTable is the shared table layout for the routes command.
var baseTable = table.New().
	Border(lipgloss.HiddenBorder()).
	StyleFunc(routesTableStyle)

// newFlightTable creates a borderless table for flight rows in the find command.
func newFlightTable(rows [][]string) *table.Table {
	return table.New().
		Border(lipgloss.HiddenBorder()).
		StyleFunc(flightTableStyle).
		Rows(rows...)
}

// flightTableStyle styles find flight tables: header dim, body plain.
// Padding is applied per cell for comfortable spacing.
func flightTableStyle(_, col int) lipgloss.Style {
	if col == 0 {
		return headerStyle.Padding(0, 2)
	}
	return lipgloss.NewStyle().Padding(0, 2)
}

// routesTableStyle styles the routes table columns: dim arrow, bold Avios.
// Padding is applied per cell for comfortable spacing.
func routesTableStyle(_, col int) lipgloss.Style {
	switch col {
	case 1:
		return arrowStyle.Padding(0, 1)
	case 3, 4:
		return numberStyle.Padding(0, 2)
	default:
		return lipgloss.NewStyle().Padding(0, 2)
	}
}

// seatStyle colours a seat count by availability. Colour is disabled
// automatically for non-TTY output or when NO_COLOR is set.
func seatStyle(seats int) lipgloss.Style {
	switch {
	case seats >= 9:
		return seatGreen
	case seats >= 5:
		return seatYellow
	case seats >= 1:
		return seatRed
	default:
		return lipgloss.NewStyle()
	}
}

func adaptiveColour(code string) compat.AdaptiveColor {
	colour := lipgloss.Color(code)
	return compat.AdaptiveColor{
		Light: colour,
		Dark:  colour,
	}
}

func cabinStyle(code string) lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(adaptiveColour(code))
}

func seatColour(code string) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(adaptiveColour(code))
}
