package theme

import (
	"fmt"
)

type RGB struct {
	R, G, B int
}

func (rgb RGB) ToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

func HexToRGB(hex string) (RGB, error) {
	if len(hex) == 6 {
		hex = "#" + hex // Add # if missing
	}
	
	if len(hex) != 7 || hex[0] != '#' {
		return RGB{}, fmt.Errorf("invalid hex color format: %s", hex)
	}

	var r, g, b int
	_, err := fmt.Sscanf(hex[1:], "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return RGB{}, fmt.Errorf("failed to parse hex color: %w", err)
	}

	return RGB{R: r, G: g, B: b}, nil
}