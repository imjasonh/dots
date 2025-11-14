package braille

import (
	"fmt"
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// Options configures the braille conversion.
type Options struct {
	Width     int   // Width in braille characters
	Height    int   // Height in braille characters
	Threshold uint8 // Brightness threshold (0-255), default 20
	Dither    bool  // Enable Floyd-Steinberg dithering
	Color     bool  // Enable ANSI color output
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

	// Step 1: Spatial quantization - resize to target dimensions
	// Each braille char is 2 pixels wide × 4 pixels tall
	targetWidth := opts.Width * 2
	targetHeight := opts.Height * 4
	resized := resize(img, targetWidth, targetHeight)

	// Apply dithering if requested
	if opts.Dither {
		resized = applyDithering(resized, opts.Threshold)
	}

	// Step 2 & 3: Brightness and color quantization
	lines := make([]string, opts.Height)

	for row := 0; row < opts.Height; row++ {
		line := ""
		for col := 0; col < opts.Width; col++ {
			// Extract 2×4 pixel block
			x0, y0 := col*2, row*4
			block := extractBlock(resized, x0, y0)

			// Brightness quantization: convert to braille character
			char := blockToBraille(block, opts.Threshold)

			// Color quantization: get ANSI color code
			if opts.Color {
				colorCode := blockToANSI(block)
				line += ansiColor(colorCode) + string(char) + ansiReset()
			} else {
				line += string(char)
			}
		}
		lines[row] = line
	}

	return lines
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

// ansiColor returns the ANSI escape sequence to set foreground color.
func ansiColor(code uint8) string {
	return fmt.Sprintf("\x1b[38;5;%dm", code)
}

// ansiReset returns the ANSI escape sequence to reset colors.
func ansiReset() string {
	return "\x1b[0m"
}

// applyDithering applies Floyd-Steinberg dithering to an image.
// This distributes quantization error to neighboring pixels for better gradient representation.
func applyDithering(img *image.RGBA, threshold uint8) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Copy image to result so we can modify it
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, img.At(x, y))
		}
	}

	// Floyd-Steinberg dithering
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			oldPixel := result.RGBAAt(x, y)

			// Convert to grayscale
			r, g, b := oldPixel.R, oldPixel.G, oldPixel.B
			luminance := uint8(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))

			// Quantize to black or white
			var newPixel uint8
			if luminance > threshold {
				newPixel = 255
			} else {
				newPixel = 0
			}

			// Calculate quantization error
			err := int(luminance) - int(newPixel)

			// Set new pixel value
			result.SetRGBA(x, y, color.RGBA{R: newPixel, G: newPixel, B: newPixel, A: 255})

			// Distribute error to neighboring pixels (Floyd-Steinberg matrix)
			// Pattern:     X   7/16
			//         3/16 5/16 1/16
			distributeError := func(dx, dy int, factor float64) {
				nx, ny := x+dx, y+dy
				if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
					oldColor := result.RGBAAt(nx, ny)
					newValue := int(oldColor.R) + int(float64(err)*factor)
					if newValue < 0 {
						newValue = 0
					}
					if newValue > 255 {
						newValue = 255
					}
					result.SetRGBA(nx, ny, color.RGBA{
						R: uint8(newValue),
						G: uint8(newValue),
						B: uint8(newValue),
						A: 255,
					})
				}
			}

			distributeError(1, 0, 7.0/16.0)
			distributeError(-1, 1, 3.0/16.0)
			distributeError(0, 1, 5.0/16.0)
			distributeError(1, 1, 1.0/16.0)
		}
	}

	return result
}
