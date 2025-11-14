package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/imjasonh/dots"
)

func main() {
	var (
		width      = flag.Int("width", 0, "Output width in characters (default: terminal width)")
		height     = flag.Int("height", 0, "Output height in characters (default: terminal height)")
		w          = flag.Int("w", 0, "Short form of -width")
		h          = flag.Int("h", 0, "Short form of -height")
		noColor    = flag.Bool("no-color", false, "Disable ANSI colors")
		background = flag.String("background", "", "Background color as hex (e.g., 'ff0000' for red, enables ANSI background)")
		threshold  = flag.Int("threshold", 20, "Brightness threshold (0-255)")
		t          = flag.Int("t", 0, "Short form of -threshold")
		frame      = flag.Bool("frame", false, "Draw a white ASCII frame around the picture")
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

	// Parse background color if provided
	var bgColor *uint8
	if *background != "" {
		ansiColor, err := dots.ParseHex(*background)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid background color: %v\n", err)
			os.Exit(1)
		}
		bgColor = &ansiColor
	}

	// Convert to dots
	lines := dots.Convert(img, dots.Options{
		Width:           *width,
		Height:          *height,
		Threshold:       uint8(*threshold),
		NoColor:         *noColor,
		BackgroundColor: bgColor,
		Frame:           *frame,
	})

	// Print output
	for _, line := range lines {
		fmt.Println(line)
	}
}
