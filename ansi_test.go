package dots

import (
	"fmt"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestANSIColorFunctions(t *testing.T) {
	for _, tt := range []struct {
		desc string
		fn   func() string
		want string
	}{
		{
			desc: "foreground color",
			fn:   func() string { return ansiFgColor(196) },
			want: "\x1b[38;5;196m",
		},
		{
			desc: "foreground and background color",
			fn:   func() string { return ansiFgBgColor(196, 21) },
			want: "\x1b[38;5;196;48;5;21m",
		},
		{
			desc: "reset",
			fn:   ansiReset,
			want: "\x1b[0m",
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBackgroundOption(t *testing.T) {
	// Create a simple red image using existing helper
	img := createTestImage(t, "test_red.png", 8, 16, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Load it back
	f, _ := os.Open(img)
	defer f.Close()
	decoded, _ := png.Decode(f)

	// Without background
	linesNoBackground := Convert(decoded, Options{
		Width:           2,
		Height:          2,
		NoColor:         false,
		BackgroundColor: nil,
	})

	// With background (red color, ANSI code 196)
	redBg := uint8(196)
	linesWithBackground := Convert(decoded, Options{
		Width:           2,
		Height:          2,
		NoColor:         false,
		BackgroundColor: &redBg,
	})

	// Check that background version includes both fg and bg codes
	for i, line := range linesWithBackground {
		if len(line) < len(linesNoBackground[i]) {
			t.Errorf("line %d: background version should be longer", i)
		}
		// Should contain ";48;5;" which is the background color prefix
		if !containsSubstring(line, ";48;5;") {
			t.Errorf("line %d: should contain background color code ;48;5;", i)
		}
	}

	// No background version should not have background codes
	for i, line := range linesNoBackground {
		if containsSubstring(line, ";48;5;") {
			t.Errorf("line %d: should not contain background color code ;48;5;", i)
		}
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBackgroundColorQuantization(t *testing.T) {
	// Test that different background colors are properly quantized and applied
	img := createTestImage(t, "test_white.png", 8, 16, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	f, _ := os.Open(img)
	defer f.Close()
	decoded, _ := png.Decode(f)

	testCases := []struct {
		desc    string
		r, g, b uint8
		wantBg  uint8
	}{
		{
			desc: "pure red",
			r:    255, g: 0, b: 0,
			wantBg: 196, // ANSI code for red
		},
		{
			desc: "pure green",
			r:    0, g: 255, b: 0,
			wantBg: 46, // ANSI code for green
		},
		{
			desc: "pure blue",
			r:    0, g: 0, b: 255,
			wantBg: 21, // ANSI code for blue
		},
		{
			desc: "black",
			r:    0, g: 0, b: 0,
			wantBg: 16, // ANSI code for black
		},
		{
			desc: "white",
			r:    255, g: 255, b: 255,
			wantBg: 231, // ANSI code for white
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Quantize the color
			bgColor := quantizeRGB(tc.r, tc.g, tc.b)

			if bgColor != tc.wantBg {
				t.Errorf("QuantizeRGB(%d, %d, %d) = %d, want %d",
					tc.r, tc.g, tc.b, bgColor, tc.wantBg)
			}

			// Test that it's used in the output
			lines := Convert(decoded, Options{
				Width:           2,
				Height:          2,
				NoColor:         false,
				BackgroundColor: &bgColor,
			})

			// Check that background code appears in output
			// Use fmt.Sprintf for proper formatting
			expectedBgCode := fmt.Sprintf(";48;5;%d", bgColor)

			found := false
			for _, line := range lines {
				if containsSubstring(line, expectedBgCode) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected background code %q not found in output", expectedBgCode)
			}
		})
	}
}

func TestBackgroundColorNil(t *testing.T) {
	// Test that nil background doesn't add background codes
	img := createTestImage(t, "test_color.png", 8, 16, color.RGBA{R: 128, G: 64, B: 200, A: 255})
	f, _ := os.Open(img)
	defer f.Close()
	decoded, _ := png.Decode(f)

	lines := Convert(decoded, Options{
		Width:           2,
		Height:          2,
		NoColor:         false,
		BackgroundColor: nil,
	})

	for i, line := range lines {
		if containsSubstring(line, ";48;5;") {
			t.Errorf("line %d: should not contain background code with nil BackgroundColor", i)
		}
		// Should still have foreground codes
		if !containsSubstring(line, "[38;5;") {
			t.Errorf("line %d: should contain foreground color code", i)
		}
	}
}
