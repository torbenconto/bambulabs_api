package utils

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
)

func IsValidGCode(line string) bool {
	line = strings.Split(line, ";")[0]
	line = strings.TrimSpace(line)

	re := regexp.MustCompile(`^[GM]\d+`)
	if line == "" || !re.MatchString(line) {
		return false
	}

	tokens := strings.Fields(line)
	for _, token := range tokens[1:] {
		paramRe := regexp.MustCompile(`^[A-Z]-?\d+(\.\d+)?$`)
		if !paramRe.MatchString(token) {
			return false
		}
	}

	return true
}

// https://stackoverflow.com/a/54200713
func ParseHexColorFast(s string) (c color.RGBA, err error) {
	// Remove the '#' if it's present
	hex := strings.TrimPrefix(s, "#")
	var r, g, b, a uint8

	// Parse the hex string based on its length
	switch len(hex) {
	case 6: // RGB format
		rVal, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		gVal, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		bVal, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		r, g, b, a = uint8(rVal), uint8(gVal), uint8(bVal), 255
	case 8: // RGBA format
		rVal, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		gVal, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		bVal, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		aVal, err := strconv.ParseUint(hex[6:8], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		r, g, b, a = uint8(rVal), uint8(gVal), uint8(bVal), uint8(aVal)
	default:
		return color.RGBA{}, fmt.Errorf("invalid hex color length: %s", hex)
	}

	return color.RGBA{R: r, G: g, B: b, A: a}, nil
}
