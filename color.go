package dots

import "fmt"

// ParseHex parses a hex color string (with or without #) and returns the ANSI 256 color code.
// Supports both 3-character shorthand (e.g., "f00") and 6-character full format (e.g., "ff0000").
func ParseHex(hex string) (uint8, error) {
	// Remove # prefix if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	// Support both 3-char and 6-char hex
	if len(hex) == 3 {
		// Expand RGB to RRGGBB
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	if len(hex) != 6 {
		return 0, fmt.Errorf("invalid hex color length: %d (expected 3 or 6)", len(hex))
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, fmt.Errorf("invalid hex color format: %w", err)
	}

	return quantizeRGB(r, g, b), nil
}
