package braille

import "testing"

func TestCalculateDimensions(t *testing.T) {
	for _, tt := range []struct {
		desc           string
		imgWidth       int
		imgHeight      int
		width          int
		height         int
		maxWidth       int
		maxHeight      int
		expectedWidth  int
		expectedHeight int
	}{
		{
			desc:           "both dimensions specified - use as-is",
			imgWidth:       100,
			imgHeight:      100,
			width:          50,
			height:         25,
			maxWidth:       80,
			maxHeight:      24,
			expectedWidth:  50,
			expectedHeight: 25,
		},
		{
			desc:           "neither dimension specified, no max - return zeros",
			imgWidth:       100,
			imgHeight:      100,
			width:          0,
			height:         0,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  0,
			expectedHeight: 0,
		},
		{
			desc:           "neither specified, use terminal as constraint (square fits width)",
			imgWidth:       100,
			imgHeight:      100,
			width:          0,
			height:         0,
			maxWidth:       80,
			maxHeight:      24,
			expectedWidth:  48, // min(80, 24*100/100*2) = min(80, 48) = 48
			expectedHeight: 24, // 48*100/100/2 = 24
		},
		{
			desc:           "neither specified, use terminal as constraint (wide image fits height)",
			imgWidth:       200,
			imgHeight:      50,
			width:          0,
			height:         0,
			maxWidth:       80,
			maxHeight:      24,
			expectedWidth:  80, // If width=80, height would be 10, fits in 24
			expectedHeight: 10, // 80*50/200/2 = 10
		},
		{
			desc:           "neither specified, use terminal as constraint (tall image fits width)",
			imgWidth:       50,
			imgHeight:      200,
			width:          0,
			height:         0,
			maxWidth:       80,
			maxHeight:      24,
			expectedWidth:  12, // 24*50/200*2 = 12
			expectedHeight: 24, // If height=24, width would be 12, fits in 80
		},
		{
			desc:           "square image, width specified",
			imgWidth:       100,
			imgHeight:      100,
			width:          20,
			height:         0,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  20,
			expectedHeight: 10, // 20 * 100/100 / 2 = 10
		},
		{
			desc:           "square image, height specified",
			imgWidth:       100,
			imgHeight:      100,
			width:          0,
			height:         10,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  20, // 10 * 100/100 * 2 = 20
			expectedHeight: 10,
		},
		{
			desc:           "wide image (4:1), width specified",
			imgWidth:       200,
			imgHeight:      50,
			width:          40,
			height:         0,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  40,
			expectedHeight: 5, // 40 * 50/200 / 2 = 5
		},
		{
			desc:           "wide image (4:1), height specified",
			imgWidth:       200,
			imgHeight:      50,
			width:          0,
			height:         10,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  80, // 10 * 200/50 * 2 = 80
			expectedHeight: 10,
		},
		{
			desc:           "tall image (1:4), width specified",
			imgWidth:       50,
			imgHeight:      200,
			width:          10,
			height:         0,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  10,
			expectedHeight: 20, // 10 * 200/50 / 2 = 20
		},
		{
			desc:           "tall image (1:4), height specified",
			imgWidth:       50,
			imgHeight:      200,
			width:          0,
			height:         40,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  20, // 40 * 50/200 * 2 = 20
			expectedHeight: 40,
		},
		{
			desc:           "very wide image (16:9), width specified",
			imgWidth:       1920,
			imgHeight:      1080,
			width:          80,
			height:         0,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  80,
			expectedHeight: 22, // 80 * 1080/1920 / 2 ≈ 22.5 → 22
		},
		{
			desc:           "very wide image (16:9), height specified",
			imgWidth:       1920,
			imgHeight:      1080,
			width:          0,
			height:         45,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  160, // 45 * 1920/1080 * 2 = 160
			expectedHeight: 45,
		},
		{
			desc:           "tiny image, width specified, minimum height",
			imgWidth:       10,
			imgHeight:      2,
			width:          10,
			height:         0,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  10,
			expectedHeight: 1, // Would be 0, clamped to 1
		},
		{
			desc:           "tiny image, height specified, minimum width",
			imgWidth:       2,
			imgHeight:      10,
			width:          0,
			height:         10,
			maxWidth:       0,
			maxHeight:      0,
			expectedWidth:  4, // 10 * 2/10 * 2 = 4
			expectedHeight: 10,
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			gotWidth, gotHeight := CalculateDimensions(tt.imgWidth, tt.imgHeight, tt.width, tt.height, tt.maxWidth, tt.maxHeight)
			if gotWidth != tt.expectedWidth || gotHeight != tt.expectedHeight {
				t.Errorf("CalculateDimensions(%d, %d, %d, %d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.imgWidth, tt.imgHeight, tt.width, tt.height, tt.maxWidth, tt.maxHeight,
					gotWidth, gotHeight, tt.expectedWidth, tt.expectedHeight)
			}
		})
	}
}

func TestCalculateDimensionsAspectRatio(t *testing.T) {
	// Test that aspect ratios are preserved correctly
	// For a square image, braille output should be 2:1 (width:height) ratio

	// Square 100x100
	w, h := CalculateDimensions(100, 100, 20, 0, 0, 0)
	if w != 20 || h != 10 {
		t.Errorf("Square with width=20: got %dx%d, want 20x10", w, h)
	}
	// Verify the pixel-level aspect ratio is maintained
	pixelWidth := w * 2
	pixelHeight := h * 4
	if pixelWidth != 40 || pixelHeight != 40 {
		t.Errorf("Square pixel dimensions: got %dx%d, want 40x40", pixelWidth, pixelHeight)
	}

	// 4:1 wide image
	w, h = CalculateDimensions(400, 100, 40, 0, 0, 0)
	if w != 40 || h != 5 {
		t.Errorf("4:1 wide with width=40: got %dx%d, want 40x5", w, h)
	}
	pixelWidth = w * 2
	pixelHeight = h * 4
	imgAspect := 400.0 / 100.0
	outputAspect := float64(pixelWidth) / float64(pixelHeight)
	if outputAspect != imgAspect {
		t.Errorf("4:1 aspect not preserved: got %.2f, want %.2f", outputAspect, imgAspect)
	}

	// 1:4 tall image
	w, h = CalculateDimensions(100, 400, 10, 0, 0, 0)
	if w != 10 || h != 20 {
		t.Errorf("1:4 tall with width=10: got %dx%d, want 10x20", w, h)
	}
	pixelWidth = w * 2
	pixelHeight = h * 4
	imgAspect = 100.0 / 400.0
	outputAspect = float64(pixelWidth) / float64(pixelHeight)
	if outputAspect != imgAspect {
		t.Errorf("1:4 aspect not preserved: got %.2f, want %.2f", outputAspect, imgAspect)
	}
}
