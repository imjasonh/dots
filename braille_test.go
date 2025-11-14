package dots

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// createTestImage creates a simple test image and saves it to testdata.
func createTestImage(t *testing.T, name string, width, height int, fillColor color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fillColor)
		}
	}

	path := "testdata/" + name
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	return path
}

// createCheckerboard creates a checkerboard pattern test image.
func createCheckerboard(t *testing.T, name string, size int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if (x/4+y/4)%2 == 0 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}

	path := "testdata/" + name
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	return path
}

// createGradient creates a horizontal gradient test image.
func createGradient(t *testing.T, name string, width, height int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			brightness := uint8(x * 255 / width)
			img.Set(x, y, color.Gray{Y: brightness})
		}
	}

	path := "testdata/" + name
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	return path
}

func TestBlockToBraille(t *testing.T) {
	for _, tt := range []struct {
		desc      string
		block     [8]color.Color
		threshold uint8
		want      rune
	}{
		{
			desc:      "all white (all dots on)",
			block:     [8]color.Color{color.White, color.White, color.White, color.White, color.White, color.White, color.White, color.White},
			threshold: 128,
			want:      '⣿', // U+28FF - all 8 dots
		},
		{
			desc:      "all black (no dots)",
			block:     [8]color.Color{color.Black, color.Black, color.Black, color.Black, color.Black, color.Black, color.Black, color.Black},
			threshold: 128,
			want:      '⠀', // U+2800 - empty braille
		},
		{
			desc:      "first dot only",
			block:     [8]color.Color{color.White, color.Black, color.Black, color.Black, color.Black, color.Black, color.Black, color.Black},
			threshold: 128,
			want:      '⠁', // U+2801 - dot 1
		},
		{
			desc:      "last dot only",
			block:     [8]color.Color{color.Black, color.Black, color.Black, color.Black, color.Black, color.Black, color.Black, color.White},
			threshold: 128,
			want:      '⢀', // U+2880 - dot 8
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got := blockToBraille(tt.block, tt.threshold)
			if got != tt.want {
				t.Errorf("blockToBraille() = %U (%c), want %U (%c)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestQuantizeRGB(t *testing.T) {
	for _, tt := range []struct {
		desc    string
		r, g, b uint8
		want    uint8
	}{
		{
			desc: "black to grayscale",
			r:    0, g: 0, b: 0,
			want: 16, // Black from RGB cube
		},
		{
			desc: "white to grayscale",
			r:    255, g: 255, b: 255,
			want: 231, // White from RGB cube
		},
		{
			desc: "pure red",
			r:    255, g: 0, b: 0,
			want: 196, // 16 + 36*5 + 6*0 + 0
		},
		{
			desc: "pure green",
			r:    0, g: 255, b: 0,
			want: 46, // 16 + 36*0 + 6*5 + 0
		},
		{
			desc: "pure blue",
			r:    0, g: 0, b: 255,
			want: 21, // 16 + 36*0 + 6*0 + 5
		},
		{
			desc: "mid gray",
			r:    128, g: 128, b: 128,
			want: 244, // Grayscale ramp
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got := quantizeRGB(tt.r, tt.g, tt.b)
			if got != tt.want {
				t.Errorf("quantizeRGB(%d, %d, %d) = %d, want %d", tt.r, tt.g, tt.b, got, tt.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	// Create test images
	_ = createTestImage(t, "white.png", 8, 16, color.White)
	_ = createTestImage(t, "black.png", 8, 16, color.Black)
	_ = createCheckerboard(t, "checkerboard.png", 16)
	_ = createGradient(t, "gradient.png", 64, 16)
	_ = createTestImage(t, "red.png", 8, 16, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	for _, tt := range []struct {
		desc     string
		imgPath  string
		opts     Options
		validate func(t *testing.T, lines []string)
	}{
		{
			desc:    "white image produces all-dots braille",
			imgPath: "testdata/white.png",
			opts:    Options{Width: 4, Height: 4, Threshold: 128, NoColor: true},
			validate: func(t *testing.T, lines []string) {
				if len(lines) != 4 {
					t.Errorf("got %d lines, want 4", len(lines))
				}
				// Each line should have 4 characters, all should be ⣿ (all dots on)
				for i, line := range lines {
					runes := []rune(line)
					if len(runes) != 4 {
						t.Errorf("line %d: got %d chars, want 4", i, len(runes))
					}
					for j, r := range runes {
						if r != '⣿' {
							t.Errorf("line %d, char %d: got %c, want ⣿", i, j, r)
						}
					}
				}
			},
		},
		{
			desc:    "black image produces empty braille",
			imgPath: "testdata/black.png",
			opts:    Options{Width: 4, Height: 4, Threshold: 128, NoColor: true},
			validate: func(t *testing.T, lines []string) {
				if len(lines) != 4 {
					t.Errorf("got %d lines, want 4", len(lines))
				}
				// Each line should have 4 characters, all should be ⠀ (no dots)
				for i, line := range lines {
					runes := []rune(line)
					if len(runes) != 4 {
						t.Errorf("line %d: got %d chars, want 4", i, len(runes))
					}
					for j, r := range runes {
						if r != '⠀' {
							t.Errorf("line %d, char %d: got %c, want ⠀", i, j, r)
						}
					}
				}
			},
		},
		{
			desc:    "checkerboard produces varied braille",
			imgPath: "testdata/checkerboard.png",
			opts:    Options{Width: 4, Height: 4, Threshold: 128, NoColor: true},
			validate: func(t *testing.T, lines []string) {
				if len(lines) != 4 {
					t.Errorf("got %d lines, want 4", len(lines))
				}
				// Should have a mix of different braille characters
				uniqueChars := make(map[rune]bool)
				for _, line := range lines {
					for _, r := range line {
						uniqueChars[r] = true
					}
				}
				if len(uniqueChars) < 2 {
					t.Errorf("checkerboard should produce at least 2 different braille chars, got %d", len(uniqueChars))
				}
			},
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			// Load image
			f, err := os.Open(tt.imgPath)
			if err != nil {
				t.Fatalf("failed to open test image: %v", err)
			}
			defer f.Close()

			img, err := png.Decode(f)
			if err != nil {
				t.Fatalf("failed to decode test image: %v", err)
			}

			// Convert
			lines := Convert(img, tt.opts)

			// Validate
			tt.validate(t, lines)
		})
	}
}

func TestResize(t *testing.T) {
	// Create a simple test image
	src := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			src.Set(x, y, color.White)
		}
	}

	// Resize to smaller dimensions
	resized := resize(src, 10, 10)

	if resized.Bounds().Dx() != 10 {
		t.Errorf("resized width = %d, want 10", resized.Bounds().Dx())
	}
	if resized.Bounds().Dy() != 10 {
		t.Errorf("resized height = %d, want 10", resized.Bounds().Dy())
	}
}
