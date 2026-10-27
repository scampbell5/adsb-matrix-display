package skywatch

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	assetpkg "example.com/adsb-matrix-display/internal/assets"
)

const FrameWidth = 64
const FrameHeight = 32

var (
	black  = color.RGBA{0, 0, 0, 255}
	white  = color.RGBA{245, 248, 255, 255}
	cyan   = color.RGBA{48, 210, 255, 255}
	yellow = color.RGBA{255, 214, 10, 255}
	green  = color.RGBA{74, 222, 128, 255}
	muted  = color.RGBA{135, 150, 170, 255}
)

func RenderFramePNG(model DisplayModel) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, FrameWidth, FrameHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)
	set := func(x, y int, c color.Color) {
		if x >= 0 && x < FrameWidth && y >= 0 && y < FrameHeight {
			img.Set(x, y, c)
		}
	}

	if model.Idle {
		line1, line2 := idleNameLines(model.DisplayName)
		drawText5x7(set, 3, 4, line1, cyan)
		drawText5x7(set, 3, 15, line2, white)
		// tiny aircraft marker at right
		drawText3x5(set, 51, 25, "ADS", muted)
	} else {
		logo, err := assetpkg.OperatorIcon(model.OperatorCode)
		if err != nil {
			return nil, err
		}
		if icon, err := png.Decode(bytes.NewReader(logo)); err == nil {
			draw.Draw(img, image.Rect(0, 0, 20, 20), icon, icon.Bounds().Min, draw.Over)
		}

		cs := strings.ToUpper(model.Callsign)
		if len(cs) > 7 {
			cs = cs[:7]
		}
		drawText5x7(set, 21, 1, cs, cyan)

		typ := strings.ToUpper(model.Type)
		if typ == "" {
			typ = "----"
		}
		if len(typ) > 4 {
			typ = typ[:4]
		}
		drawText3x5(set, 22, 11, typ, white)
		drawText3x5(set, 50, 11, model.Direction, green)

		// Three compact metric blocks: distance, altitude, speed.
		drawText3x5(set, 0, 23, model.Distance, yellow)
		drawText3x5(set, 22, 23, compactAltitude(model.Altitude), white)
		drawText3x5(set, 44, 23, model.Speed, white)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func compactAltitude(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "GND" || s == "---" {
		return s
	}
	n := strings.TrimSuffix(s, "FT")
	var feet int
	if _, err := fmt.Sscanf(n, "%d", &feet); err != nil {
		return n
	}
	if feet >= 10000 {
		return fmt.Sprintf("%.0fKF", float64(feet)/1000)
	}
	if feet >= 1000 {
		return fmt.Sprintf("%.1fKF", float64(feet)/1000)
	}
	return fmt.Sprintf("%dF", feet)
}

func idleNameLines(name string) (string, string) {
	s := strings.ToUpper(strings.TrimSpace(name))
	if s == "" {
		s = "ADS-B MATRIX"
	}
	fields := strings.Fields(s)
	if len(fields) >= 2 {
		line1 := strings.Join(fields[:len(fields)-1], " ")
		line2 := fields[len(fields)-1]
		return truncatePixelLine(line1, 9), truncatePixelLine(line2, 9)
	}
	if len(s) <= 9 {
		return truncatePixelLine(s, 9), "DISPLAY"
	}
	return truncatePixelLine(s[:len(s)/2], 9), truncatePixelLine(s[len(s)/2:], 9)
}

func truncatePixelLine(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = s[:max]
	}
	return s
}
