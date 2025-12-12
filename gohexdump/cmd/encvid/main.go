package main

import (
	"flag"
	"io"
	"log"
	"os"

	"post6.net/gohexdump/internal/screen"
)

var (
	width  int
	height int
)

func init() {
	flag.IntVar(&width, "width", 1280, "input video width in pixels")
	flag.IntVar(&height, "height", 720, "input video height in pixels")
}

// buildPositions builds a mapping from segment index -> pixel index in the input frame.
//
// For each segment i:
//   - We get its physical coordinate via SegmentCoord(i)
//   - Normalize all segment coordinates into [0..width-1] × [0..height-1]
//   - Store the linear pixel index x + y*width in positions[i]
//   - For the 16th segment in each digit (ix & 0xF == 0xF) we store -1 (unused)
func buildPositions(w, h int) []int {
	hex := screen.NewHexScreen()

	segCount := hex.SegmentCount()
	if segCount <= 0 {
		log.Fatalf("SegmentCount() returned %d", segCount)
	}

	positions := make([]int, segCount)

	// Find bounding box of all segment coordinates
	const big = 1e9
	lx, ly := big, big
	hx, hy := -big, -big

	for i := 0; i < segCount; i++ {
		c := hex.SegmentCoord(i)
		if c.X < lx {
			lx = c.X
		}
		if c.X > hx {
			hx = c.X
		}
		if c.Y < ly {
			ly = c.Y
		}
		if c.Y > hy {
			hy = c.Y
		}
	}

	if hx <= lx || hy <= ly {
		log.Fatalf("invalid segment coordinate bounds: lx=%f hx=%f ly=%f hy=%f", lx, hx, ly, hy)
	}

	fx := float64(w-1) / (hx - lx)
	fy := float64(h-1) / (hy - ly)

	for i := 0; i < segCount; i++ {
		// Skip the 16th segment in each digit (center/unused)
		if i&0xF == 0xF {
			positions[i] = -1
			continue
		}

		c := hex.SegmentCoord(i)

		x := int((c.X - lx) * fx)
		y := int((c.Y - ly) * fy)

		// Clamp to image bounds
		if x < 0 {
			x = 0
		}
		if x >= w {
			x = w - 1
		}
		if y < 0 {
			y = 0
		}
		if y >= h {
			y = h - 1
		}

		positions[i] = x + y*w
	}

	return positions
}

func main() {
	flag.Parse()

	if width <= 0 || height <= 0 {
		log.Fatalf("invalid dimensions: width=%d height=%d", width, height)
	}

	positions := buildPositions(width, height)

	inFrameSize := width * height
	inframe := make([]byte, inFrameSize)
	outframe := make([]byte, len(positions))

	for {
		_, err := io.ReadFull(os.Stdin, inframe)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// Normal end of stream
				break
			}
			log.Fatalf("reading input frame: %v", err)
		}

		for i, p := range positions {
			if p >= 0 && p < len(inframe) {
				outframe[i] = inframe[p]
			} else {
				// Out-of-range or unused segment -> black
				outframe[i] = 0
			}
		}

		if _, err := os.Stdout.Write(outframe); err != nil {
			log.Fatalf("writing output frame: %v", err)
		}
	}
}
