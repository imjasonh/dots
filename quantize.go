package dots

import "math"

// quantizeRGB maps an RGB color to the nearest ANSI 256 color code.
// ANSI 256 color palette:
//   - 0-15: System colors (we avoid these for consistency)
//   - 16-231: 6×6×6 RGB cube (216 colors)
//   - 232-255: 24-step grayscale ramp
func quantizeRGB(r, g, b uint8) uint8 {
	// Check if color is grayscale (low saturation)
	if isGrayscale(r, g, b) {
		return quantizeGrayscale(r, g, b)
	}

	// Quantize to 6×6×6 RGB cube
	return quantizeRGBCube(r, g, b)
}

// isGrayscale determines if a color has low enough saturation to be treated as grayscale.
func isGrayscale(r, g, b uint8) bool {
	max := max3(r, g, b)
	min := min3(r, g, b)
	// If the difference between max and min channels is small, treat as grayscale
	return max-min < 10
}

// quantizeGrayscale maps a grayscale color to the 24-step grayscale ramp (232-255).
func quantizeGrayscale(r, g, b uint8) uint8 {
	// Calculate average brightness
	avg := (uint16(r) + uint16(g) + uint16(b)) / 3

	// Map to 24-step grayscale (232-255)
	// Grayscale range: 8 to 238 (in steps of ~10)
	if avg < 8 {
		return 16 // Black from RGB cube
	}
	if avg >= 238 {
		return 231 // White from RGB cube
	}

	// Map [8, 237] to [0, 23]
	step := (avg - 8) * 24 / 229
	if step > 23 {
		step = 23
	}
	return 232 + uint8(step)
}

// quantizeRGBCube maps an RGB color to the 6×6×6 RGB cube (colors 16-231).
func quantizeRGBCube(r, g, b uint8) uint8 {
	// Each channel is quantized to 6 levels: 0, 1, 2, 3, 4, 5
	// The actual values are: 0, 95, 135, 175, 215, 255
	r6 := quantizeChannel(r)
	g6 := quantizeChannel(g)
	b6 := quantizeChannel(b)

	// Formula: 16 + 36*r + 6*g + b
	return 16 + 36*r6 + 6*g6 + b6
}

// quantizeChannel maps a single color channel [0, 255] to a 6-level value [0, 5].
// Uses nearest-neighbor quantization with the actual palette values.
func quantizeChannel(c uint8) uint8 {
	// ANSI 256 color RGB cube values for each level
	levels := []uint8{0, 95, 135, 175, 215, 255}

	// Find nearest level
	minDist := uint8(255)
	nearest := uint8(0)

	for i, level := range levels {
		dist := absDiff(c, level)
		if dist < minDist {
			minDist = dist
			nearest = uint8(i)
		}
	}

	return nearest
}

// absDiff returns the absolute difference between two uint8 values.
func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

// max3 returns the maximum of three uint8 values.
func max3(a, b, c uint8) uint8 {
	return uint8(math.Max(float64(a), math.Max(float64(b), float64(c))))
}

// min3 returns the minimum of three uint8 values.
func min3(a, b, c uint8) uint8 {
	return uint8(math.Min(float64(a), math.Min(float64(b), float64(c))))
}
