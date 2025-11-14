package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/imjasonh/dots/braille"
	"golang.org/x/term"
)

func main() {
	var (
		width     = flag.Int("width", 0, "Output width in characters (default: terminal width)")
		height    = flag.Int("height", 0, "Output height in characters (default: terminal height)")
		w         = flag.Int("w", 0, "Short form of -width")
		h         = flag.Int("h", 0, "Short form of -height")
		noColor   = flag.Bool("no-color", false, "Disable ANSI colors")
		threshold = flag.Int("threshold", 20, "Brightness threshold (0-255)")
		t         = flag.Int("t", 0, "Short form of -threshold")
		dither    = flag.Bool("dither", false, "Enable Floyd-Steinberg dithering")
	)

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <image>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	imagePath := flag.Arg(0)

	// Resolve short flags
	if *w > 0 {
		width = w
	}
	if *h > 0 {
		height = h
	}
	if *t > 0 {
		threshold = t
	}

	// Validate threshold
	if *threshold < 0 || *threshold > 255 {
		fmt.Fprintf(os.Stderr, "Error: threshold must be between 0 and 255\n")
		os.Exit(1)
	}

	// Load image to get dimensions for aspect ratio calculation
	f, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open image: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to decode image: %v\n", err)
		os.Exit(1)
	}

	// Get terminal dimensions for use as max constraints
	fd := int(os.Stdout.Fd())
	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		// Default to reasonable values if terminal size can't be determined
		termWidth, termHeight = 80, 24
	}

	// Calculate dimensions based on what was specified
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	*width, *height = braille.CalculateDimensions(imgWidth, imgHeight, *width, *height, termWidth, termHeight)

	// Fallback if still zero (shouldn't happen with terminal dimensions)
	if *width == 0 {
		*width = termWidth
	}
	if *height == 0 {
		*height = termHeight
	}

	// Convert to braille
	opts := braille.Options{
		Width:     *width,
		Height:    *height,
		Threshold: uint8(*threshold),
		Dither:    *dither,
		Color:     !*noColor,
	}

	lines := braille.Convert(img, opts)

	// Print output
	for _, line := range lines {
		fmt.Println(line)
	}
}
