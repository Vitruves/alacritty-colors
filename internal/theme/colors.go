package theme

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
)

type HSL struct {
	H, S, L float64
}

type RGB struct {
	R, G, B int
}

func (rgb RGB) ToHSL() HSL {
	r, g, b := float64(rgb.R)/255.0, float64(rgb.G)/255.0, float64(rgb.B)/255.0

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)

	var h, s, l float64
	l = (max + min) / 2

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h /= 6
	}

	return HSL{H: h, S: s, L: l}
}

func hue2rgb(p, q, t float64) float64 {
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

func (hsl HSL) ToRGB() RGB {
	var r, g, b float64

	if hsl.S == 0 {
		r = hsl.L
		g = hsl.L
		b = hsl.L
	} else {
		var q float64
		if hsl.L < 0.5 {
			q = hsl.L * (1 + hsl.S)
		} else {
			q = hsl.L + hsl.S - hsl.L*hsl.S
		}

		p := 2*hsl.L - q
		r = hue2rgb(p, q, hsl.H+1.0/3.0)
		g = hue2rgb(p, q, hsl.H)
		b = hue2rgb(p, q, hsl.H-1.0/3.0)
	}

	return RGB{
		R: int(r * 255),
		G: int(g * 255),
		B: int(b * 255),
	}
}

func (rgb RGB) ToHex() string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

func HexToRGB(hex string) (RGB, error) {
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

func randomFloat() float64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return float64(n.Int64()) / 1000000.0
}

func randomInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}

// Color scheme generation functions
func GenerateComplementaryColors(baseHue float64) []float64 {
	return []float64{
		baseHue,
		math.Mod(baseHue+0.5, 1.0), // Complementary
	}
}

func GenerateTriadicColors(baseHue float64) []float64 {
	return []float64{
		baseHue,
		math.Mod(baseHue+1.0/3.0, 1.0),
		math.Mod(baseHue+2.0/3.0, 1.0),
	}
}

func GenerateAnalogousColors(baseHue float64) []float64 {
	return []float64{
		math.Mod(baseHue-1.0/12.0, 1.0),
		baseHue,
		math.Mod(baseHue+1.0/12.0, 1.0),
		math.Mod(baseHue+2.0/12.0, 1.0),
	}
}

func GenerateMonochromaticColors(baseHue, baseSat float64) []HSL {
	lightnesses := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}
	var colors []HSL

	for _, l := range lightnesses {
		colors = append(colors, HSL{H: baseHue, S: baseSat, L: l})
	}

	return colors
}

// Perceptual color functions
func GetLuminance(rgb RGB) float64 {
	// Convert to linear RGB
	toLinear := func(c float64) float64 {
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}

	r := toLinear(float64(rgb.R) / 255.0)
	g := toLinear(float64(rgb.G) / 255.0)
	b := toLinear(float64(rgb.B) / 255.0)

	return 0.2126*r + 0.7152*g + 0.0722*b
}

func GetContrastRatio(color1, color2 RGB) float64 {
	lum1 := GetLuminance(color1)
	lum2 := GetLuminance(color2)

	lighter := math.Max(lum1, lum2)
	darker := math.Min(lum1, lum2)

	return (lighter + 0.05) / (darker + 0.05)
}

func EnsureContrast(foreground, background RGB, minRatio float64) RGB {
	ratio := GetContrastRatio(foreground, background)
	if ratio >= minRatio {
		return foreground
	}

	// Adjust lightness to meet contrast requirement
	fgHSL := foreground.ToHSL()
	bgLum := GetLuminance(background)

	// Try making foreground lighter or darker
	for i := 0; i < 100; i++ {
		if bgLum > 0.5 {
			// Dark background, make foreground lighter
			fgHSL.L = math.Min(1.0, fgHSL.L+0.01)
		} else {
			// Light background, make foreground darker
			fgHSL.L = math.Max(0.0, fgHSL.L-0.01)
		}

		newFg := fgHSL.ToRGB()
		if GetContrastRatio(newFg, background) >= minRatio {
			return newFg
		}
	}

	return foreground
}
