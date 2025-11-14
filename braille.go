package dots

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"golang.org/x/image/draw"
	"golang.org/x/term"
)

// getTerminalSize returns the current terminal dimensions.
// Returns 80x24 as fallback if terminal size cannot be determined.
func getTerminalSize() (int, int) {
	// Try to get terminal size from stdout
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		if width, height, err := term.GetSize(fd); err == nil {
			return width, height
		}
	}
	// Fallback to reasonable defaults
	return 80, 24
}

// Options configures the braille conversion.
type Options struct {
	Width           int    // Width in braille characters
	Height          int    // Height in braille characters
	Threshold       uint8  // Brightness threshold (0-255), default 20
	NoColor         bool   // Disable ANSI color output
	BackgroundColor *uint8 // Background color for ANSI output (nil = no background)
	Frame           bool   // Draw a white ASCII frame around the picture
}

// CalculateDimensions calculates output dimensions maintaining aspect ratio.
// If both width and height are specified, returns them unchanged.
// If only width is specified, calculates height from image aspect ratio.
// If only height is specified, calculates width from image aspect ratio.
// If neither is specified, uses maxWidth and maxHeight as constraints while maintaining aspect ratio.
//
// The calculation accounts for braille characters being 2 pixels wide × 4 pixels tall.
func CalculateDimensions(imgWidth, imgHeight, width, height, maxWidth, maxHeight int) (int, int) {
	if width > 0 && height > 0 {
		// Both specified, use as-is
		return width, height
	}

	if width > 0 && height == 0 {
		// Only width specified, calculate height to maintain aspect ratio
		// width chars = width*2 pixels wide
		// To maintain aspect: height pixels = width*2 * (imgHeight/imgWidth)
		// height chars = height pixels / 4 = width*2*(imgHeight/imgWidth)/4 = width*imgHeight/imgWidth/2
		height = int(float64(width) * float64(imgHeight) / float64(imgWidth) / 2.0)
		if height == 0 {
			height = 1
		}
		return width, height
	}

	if height > 0 && width == 0 {
		// Only height specified, calculate width to maintain aspect ratio
		// height chars = height*4 pixels tall
		// To maintain aspect: width pixels = height*4 * (imgWidth/imgHeight)
		// width chars = width pixels / 2 = height*4*(imgWidth/imgHeight)/2 = height*imgWidth/imgHeight*2
		width = int(float64(height) * float64(imgWidth) / float64(imgHeight) * 2.0)
		if width == 0 {
			width = 1
		}
		return width, height
	}

	// Neither specified - use maxWidth/maxHeight as constraints and maintain aspect ratio
	if maxWidth > 0 && maxHeight > 0 {
		// Calculate what dimensions would be if we used maxWidth
		widthConstrained := maxWidth
		heightForWidth := int(float64(widthConstrained) * float64(imgHeight) / float64(imgWidth) / 2.0)

		// Calculate what dimensions would be if we used maxHeight
		heightConstrained := maxHeight
		widthForHeight := int(float64(heightConstrained) * float64(imgWidth) / float64(imgHeight) * 2.0)

		// Use whichever fits within both constraints
		if heightForWidth <= maxHeight {
			// Width-constrained version fits
			return widthConstrained, heightForWidth
		}
		// Height-constrained version fits
		return widthForHeight, heightConstrained
	}

	// No constraints at all
	return 0, 0
}

// Convert converts an image to braille representation.
// Returns a slice of strings, one per line of output.
func Convert(img image.Image, opts Options) []string {
	// Set defaults
	if opts.Threshold == 0 {
		opts.Threshold = 20
	}

	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		opts.NoColor = true
	}

	// Calculate dimensions if not both specified
	if opts.Width == 0 || opts.Height == 0 {
		bounds := img.Bounds()
		imgWidth := bounds.Dx()
		imgHeight := bounds.Dy()

		// Get terminal dimensions as constraints
		termWidth, termHeight := getTerminalSize()

		// If frame is enabled, reduce available space by 2 (1 on each side)
		if opts.Frame {
			termWidth -= 2
			termHeight -= 2
		}

		// CalculateDimensions handles all cases:
		// - Both zero: uses terminal as constraint with aspect ratio
		// - Only width: calculates height from aspect
		// - Only height: calculates width from aspect
		opts.Width, opts.Height = CalculateDimensions(imgWidth, imgHeight, opts.Width, opts.Height, termWidth, termHeight)
	} else if opts.Frame {
		// If dimensions were explicitly specified, reduce them for the frame
		opts.Width -= 2
		opts.Height -= 2
		if opts.Width < 1 {
			opts.Width = 1
		}
		if opts.Height < 1 {
			opts.Height = 1
		}
	}

	// Step 1: Spatial quantization - resize to target dimensions
	// Each braille char is 2 pixels wide × 4 pixels tall
	targetWidth := opts.Width * 2
	targetHeight := opts.Height * 4
	resized := resize(img, targetWidth, targetHeight)

	// Step 2 & 3: Brightness and color quantization
	brailleLines := make([]string, opts.Height)

	for row := 0; row < opts.Height; row++ {
		line := ""
		for col := 0; col < opts.Width; col++ {
			// Extract 2×4 pixel block
			x0, y0 := col*2, row*4
			block := extractBlock(resized, x0, y0)

			// Brightness quantization: convert to braille character
			char := blockToBraille(block, opts.Threshold)

			// Color quantization: get ANSI color codes
			if !opts.NoColor {
				fgColor := blockToANSI(block)
				if opts.BackgroundColor != nil {
					line += ansiFgBgColor(fgColor, *opts.BackgroundColor) + string(char) + ansiReset()
				} else {
					line += ansiFgColor(fgColor) + string(char) + ansiReset()
				}
			} else {
				line += string(char)
			}
		}
		brailleLines[row] = line
	}

	// Add frame if requested
	if opts.Frame {
		return addFrame(brailleLines, opts.NoColor)
	}

	return brailleLines
}

// resize scales an image to the target dimensions using high-quality interpolation.
func resize(img image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	return dst
}

// extractBlock extracts a 2×4 pixel block from an image at the given position.
func extractBlock(img *image.RGBA, x0, y0 int) [8]color.Color {
	var block [8]color.Color
	bounds := img.Bounds()

	// Standard braille dot numbering:
	// 0 3    (pixels at x0, x0+1)
	// 1 4    (rows y0, y0+1, y0+2, y0+3)
	// 2 5
	// 6 7

	positions := [][2]int{
		{x0, y0}, {x0, y0 + 1}, {x0, y0 + 2}, {x0 + 1, y0},
		{x0, y0 + 3}, {x0 + 1, y0 + 1}, {x0 + 1, y0 + 2}, {x0 + 1, y0 + 3},
	}

	for i, pos := range positions {
		x, y := pos[0], pos[1]
		if x < bounds.Max.X && y < bounds.Max.Y {
			block[i] = img.At(x, y)
		} else {
			block[i] = color.Black
		}
	}

	return block
}

// blockToBraille converts a 2×4 pixel block to a braille character.
// Each pixel's brightness is compared to the threshold to determine if the dot is on.
func blockToBraille(block [8]color.Color, threshold uint8) rune {
	var pattern uint8

	for i, c := range block {
		// Convert to grayscale using perceived luminance
		r, g, b, _ := c.RGBA()
		// RGBA() returns values in [0, 65535], convert to [0, 255]
		r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
		luminance := uint8(0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8))

		// Apply threshold: bright pixels turn on dots
		if luminance > threshold {
			pattern |= (1 << i)
		}
	}

	// Unicode braille pattern base is U+2800
	return rune(0x2800 + int(pattern))
}

// blockToANSI determines the dominant color of a block and returns the nearest ANSI 256 color code.
func blockToANSI(block [8]color.Color) uint8 {
	// Calculate average color of the block
	var rSum, gSum, bSum uint32
	for _, c := range block {
		r, g, b, _ := c.RGBA()
		rSum += r
		gSum += g
		bSum += b
	}

	// Average and convert to 8-bit
	r := uint8((rSum / 8) >> 8)
	g := uint8((gSum / 8) >> 8)
	b := uint8((bSum / 8) >> 8)

	return quantizeRGB(r, g, b)
}

// ansiFgColor returns the ANSI escape sequence to set foreground color.
func ansiFgColor(code uint8) string {
	return fmt.Sprintf("\x1b[38;5;%dm", code)
}

// ansiFgBgColor returns the ANSI escape sequence to set both foreground and background colors.
func ansiFgBgColor(fgCode, bgCode uint8) string {
	return fmt.Sprintf("\x1b[38;5;%d;48;5;%dm", fgCode, bgCode)
}

// ansiReset returns the ANSI escape sequence to reset colors.
func ansiReset() string {
	return "\x1b[0m"
}

// addFrame wraps the braille lines with a white ASCII frame.
func addFrame(lines []string, noColor bool) []string {
	if len(lines) == 0 {
		return lines
	}

	// Calculate the width of the content (without ANSI codes if present)
	width := 0
	if !noColor {
		// Count visible characters by stripping ANSI codes
		width = visibleWidth(lines[0])
	} else {
		width = len(lines[0])
	}

	// White color code (for frame)
	whiteColor := "\x1b[38;5;15m"
	reset := ""
	if !noColor {
		reset = ansiReset()
	} else {
		whiteColor = ""
	}

	// Build the frame
	result := make([]string, len(lines)+2)

	// Top border
	result[0] = whiteColor + "┌" + repeatString("─", width) + "┐" + reset

	// Content with side borders
	for i, line := range lines {
		result[i+1] = whiteColor + "│" + reset + line + whiteColor + "│" + reset
	}

	// Bottom border
	result[len(result)-1] = whiteColor + "└" + repeatString("─", width) + "┘" + reset

	return result
}

// visibleWidth counts the visible characters in a string, ignoring ANSI escape codes.
func visibleWidth(s string) int {
	width := 0
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		} else if inEscape {
			if r == 'm' {
				inEscape = false
			}
		} else {
			width++
		}
	}
	return width
}

// repeatString repeats a string n times.
func repeatString(s string, n int) string {
	result := ""
	for range n {
		result += s
	}
	return result
}
