package dots

import "testing"

func TestParseHex(t *testing.T) {
	for _, tt := range []struct {
		desc     string
		hex      string
		wantANSI uint8
		wantErr  bool
	}{
		{
			desc:     "red - 6-char hex",
			hex:      "ff0000",
			wantANSI: 196, // Red in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "red - with # prefix",
			hex:      "#ff0000",
			wantANSI: 196,
			wantErr:  false,
		},
		{
			desc:     "red - 3-char shorthand",
			hex:      "f00",
			wantANSI: 196,
			wantErr:  false,
		},
		{
			desc:     "red - 3-char with #",
			hex:      "#f00",
			wantANSI: 196,
			wantErr:  false,
		},
		{
			desc:     "blue",
			hex:      "0000ff",
			wantANSI: 21, // Blue in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "blue - short",
			hex:      "#00f",
			wantANSI: 21,
			wantErr:  false,
		},
		{
			desc:     "green",
			hex:      "00ff00",
			wantANSI: 46, // Green in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "cyan",
			hex:      "00ffff",
			wantANSI: 51, // Cyan in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "cyan - short",
			hex:      "0ff",
			wantANSI: 51,
			wantErr:  false,
		},
		{
			desc:     "yellow",
			hex:      "ffff00",
			wantANSI: 226, // Yellow in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "magenta",
			hex:      "ff00ff",
			wantANSI: 201, // Magenta in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "black",
			hex:      "000000",
			wantANSI: 16, // Black in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "white",
			hex:      "ffffff",
			wantANSI: 231, // White in ANSI 256
			wantErr:  false,
		},
		{
			desc:     "gray",
			hex:      "808080",
			wantANSI: 244, // Gray in ANSI 256 grayscale ramp
			wantErr:  false,
		},
		{
			desc:    "invalid length (too short)",
			hex:     "ff",
			wantErr: true,
		},
		{
			desc:    "invalid length (too long)",
			hex:     "ff00001",
			wantErr: true,
		},
		{
			desc:    "invalid characters",
			hex:     "gggggg",
			wantErr: true,
		},
		{
			desc:    "empty string",
			hex:     "",
			wantErr: true,
		},
		{
			desc:    "partial invalid",
			hex:     "ff00gg",
			wantErr: true,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			ansiCode, err := ParseHex(tt.hex)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseHex(%q) expected error, got nil", tt.hex)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseHex(%q) unexpected error: %v", tt.hex, err)
				return
			}

			if ansiCode != tt.wantANSI {
				t.Errorf("ParseHex(%q) = %d, want %d", tt.hex, ansiCode, tt.wantANSI)
			}
		})
	}
}

func TestParseHexConsistency(t *testing.T) {
	// Test that different formats for the same color produce the same ANSI code
	testCases := []struct {
		desc     string
		variants []string
		wantANSI uint8
	}{
		{
			desc:     "red variants",
			variants: []string{"ff0000", "#ff0000", "f00", "#f00"},
			wantANSI: 196,
		},
		{
			desc:     "blue variants",
			variants: []string{"0000ff", "#0000ff", "00f", "#00f"},
			wantANSI: 21,
		},
		{
			desc:     "cyan variants",
			variants: []string{"00ffff", "#00ffff", "0ff", "#0ff"},
			wantANSI: 51,
		},
		{
			desc:     "white variants",
			variants: []string{"ffffff", "#ffffff", "fff", "#fff"},
			wantANSI: 231,
		},
		{
			desc:     "black variants",
			variants: []string{"000000", "#000000", "000", "#000"},
			wantANSI: 16,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			for _, hex := range tc.variants {
				ansiCode, err := ParseHex(hex)
				if err != nil {
					t.Errorf("ParseHex(%q) error: %v", hex, err)
					continue
				}
				if ansiCode != tc.wantANSI {
					t.Errorf("ParseHex(%q) = %d, want %d", hex, ansiCode, tc.wantANSI)
				}
			}
		})
	}
}
