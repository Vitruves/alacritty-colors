package tui

import (
	"math"

	"github.com/vitruves/alacritty-colors/internal/theme"
)

// HSL represents a color in HSL color space
type HSL struct {
	H float64 // Hue (0-360)
	S float64 // Saturation (0-1)
	L float64 // Lightness (0-1)
}

// RGBToHSL converts RGB to HSL color space
func RGBToHSL(r, g, b int) HSL {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	maxVal := math.Max(math.Max(rf, gf), bf)
	minVal := math.Min(math.Min(rf, gf), bf)
	diff := maxVal - minVal

	// Lightness
	l := (maxVal + minVal) / 2.0

	var h, s float64

	if diff == 0 {
		h = 0 // Achromatic
		s = 0
	} else {
		// Saturation
		if l > 0.5 {
			s = diff / (2.0 - maxVal - minVal)
		} else {
			s = diff / (maxVal + minVal)
		}

		// Hue
		switch maxVal {
		case rf:
			h = (gf-bf)/diff + (func() float64 {
				if gf < bf {
					return 6.0
				}
				return 0.0
			})()
		case gf:
			h = (bf-rf)/diff + 2.0
		case bf:
			h = (rf-gf)/diff + 4.0
		}
		h /= 6.0
	}

	return HSL{H: h * 360, S: s, L: l}
}

// HSLToRGB converts HSL to RGB color space
func HSLToRGB(hsl HSL) (int, int, int) {
	h := hsl.H / 360.0
	s := hsl.S
	l := hsl.L

	var r, g, b float64

	if s == 0 {
		r = l // Achromatic
		g = l
		b = l
	} else {
		hue2rgb := func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hue2rgb(p, q, h+1.0/3.0)
		g = hue2rgb(p, q, h)
		b = hue2rgb(p, q, h-1.0/3.0)
	}

	return int(r*255 + 0.5), int(g*255 + 0.5), int(b*255 + 0.5)
}

// AdjustBrightness adjusts RGB values uniformly for brightness control
func AdjustBrightness(rgb theme.RGB, increase bool) theme.RGB {
	adjustment := BrightnessStep
	if !increase {
		adjustment = -adjustment
	}

	return theme.RGB{
		R: clampInt(rgb.R+adjustment, 0, 255),
		G: clampInt(rgb.G+adjustment, 0, 255),
		B: clampInt(rgb.B+adjustment, 0, 255),
	}
}

// AdjustHue adjusts the hue component in HSL space
func AdjustHue(rgb theme.RGB, increase bool) theme.RGB {
	hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)

	// Adjust hue
	if increase {
		hsl.H += HueStep
		if hsl.H >= 360 {
			hsl.H -= 360
		}
	} else {
		hsl.H -= HueStep
		if hsl.H < 0 {
			hsl.H += 360
		}
	}

	// Ensure minimum saturation for visible hue changes
	if hsl.S < MinSaturation {
		hsl.S = MinSaturation
	}

	// Convert back to RGB
	r, g, b := HSLToRGB(hsl)

	return theme.RGB{
		R: clampInt(r, 0, 255),
		G: clampInt(g, 0, 255),
		B: clampInt(b, 0, 255),
	}
}

// AdjustSaturation adjusts the saturation component in HSL space
func AdjustSaturation(rgb theme.RGB, increase bool) theme.RGB {
	hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)

	// Adjust saturation
	if increase {
		hsl.S = math.Min(1.0, hsl.S+SaturationStep)
	} else {
		hsl.S = math.Max(0.0, hsl.S-SaturationStep)
	}

	// Convert back to RGB
	r, g, b := HSLToRGB(hsl)

	return theme.RGB{
		R: clampInt(r, 0, 255),
		G: clampInt(g, 0, 255),
		B: clampInt(b, 0, 255),
	}
}

// AdjustLightness adjusts the lightness component in HSL space
func AdjustLightness(rgb theme.RGB, increase bool) theme.RGB {
	hsl := RGBToHSL(rgb.R, rgb.G, rgb.B)

	// Adjust lightness
	step := 0.05
	if increase {
		hsl.L = math.Min(1.0, hsl.L+step)
	} else {
		hsl.L = math.Max(0.0, hsl.L-step)
	}

	// Convert back to RGB
	r, g, b := HSLToRGB(hsl)

	return theme.RGB{
		R: clampInt(r, 0, 255),
		G: clampInt(g, 0, 255),
		B: clampInt(b, 0, 255),
	}
}

// clampInt clamps an integer value between min and max
func clampInt(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}
