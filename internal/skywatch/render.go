package skywatch

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"

	assetpkg "adsb-matrix-display/internal/assets"
)

const FrameWidth = 64
const FrameHeight = 32

const operatorIconSize = 16
const unknownIconSize = 14

var (
	black   = color.RGBA{0, 0, 0, 255}
	white   = color.RGBA{245, 248, 255, 255}
	cyan    = color.RGBA{48, 210, 255, 255}
	green   = color.RGBA{72, 232, 146, 255}
	amber   = color.RGBA{255, 190, 72, 255}
	orange  = color.RGBA{255, 128, 58, 255}
	magenta = color.RGBA{255, 72, 164, 255}
	violet  = color.RGBA{178, 112, 255, 255}
)

// carrierPalette uses bright, LED-friendly approximations of carrier identity
// colors. They are intentionally tuned for readability on black rather than
// attempting exact brand-color reproduction. Prefix-only entries let the
// callsign pick up a carrier identity even before we add a dedicated logo.
var carrierPalette = map[string]color.RGBA{
	"DAL": {245, 55, 78, 255},
	"UAL": {70, 155, 255, 255},
	"AAL": {100, 205, 255, 255},
	"SWA": {70, 150, 255, 255},
	"JBU": {60, 190, 255, 255},
	"ASA": {90, 180, 255, 255},
	"NKS": {255, 224, 48, 255},
	"FFT": {70, 225, 120, 255},
	"AAY": {255, 150, 55, 255},
	"RPA": {205, 110, 255, 255},
	"EDV": {160, 120, 255, 255},
	"ENY": {105, 205, 255, 255},
	"JIA": {255, 86, 86, 255},
	"SKW": {72, 210, 255, 255},
	"AWI": {125, 185, 255, 255},
	"ACA": {255, 70, 70, 255},
	"KAP": {58, 198, 255, 255}, // Cape Air
	"FDX": {190, 105, 255, 255},
	"UPS": {255, 188, 60, 255},
}

func RenderFramePNG(model DisplayModel) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, FrameWidth, FrameHeight))
	draw.Draw(img, img.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)
	set := func(x, y int, c color.Color) {
		if x >= 0 && x < FrameWidth && y >= 0 && y < FrameHeight {
			img.Set(x, y, c)
		}
	}

	if model.Idle {
		if err := drawOperatorIcon(img, "UNK", 26, 0, 12); err != nil {
			return nil, err
		}
		line1, line2 := idleLines(model.Callsign)
		drawCentered5x7(set, 13, line1, cyan)
		drawCentered5x7(set, 22, line2, white)
	} else {
		// Keep artwork compact so the panel reads as an instrument rather than a logo card.
		// Unknown/general-aviation aircraft use the UFO at 14x14; known operators get
		// a slightly larger 16x16 mark.
		iconSize := operatorIconSize
		iconX, iconY := 0, 0
		if model.OperatorKind == OperatorUnknown {
			iconSize = unknownIconSize
			iconX, iconY = 1, 1
		}
		if err := drawOperatorIcon(img, model.OperatorCode, iconX, iconY, iconSize); err != nil {
			return nil, err
		}

		// Callsign is the hero, but starts immediately after the compact artwork instead
		// of floating on the far right. Eight characters still fit exactly at x=17.
		cs := strings.ToUpper(strings.TrimSpace(model.Callsign))
		if len(cs) > 8 {
			cs = cs[:8]
		}
		csX := 17
		if textWidth5x7(cs) > FrameWidth-csX {
			csX = FrameWidth - textWidth5x7(cs)
			if csX < 16 {
				csX = 16
			}
		}
		drawText5x7(set, csX, 1, cs, callsignColor(model))

		// Keep the secondary identity line intentionally minimal. The operator logo
		// already identifies a known airline/branch, so only show the aircraft type
		// when we actually know it. Unknown aircraft leave this row empty.
		typ := strings.ToUpper(strings.TrimSpace(model.Type))
		if len(typ) > 4 {
			typ = typ[:4]
		}
		if typ != "" {
			drawText3x5(set, csX, 10, typ, white)
		}

		// Navigation stays one compact sentence, while its color acts as a proximity
		// level: tracked -> nearby -> close -> very close -> overhead. The arrow
		// shares the same color so the whole navigation row reads as one signal.
		distance := compactDistanceValue(model.Distance)
		distanceText := distance + " NM"
		distanceColor := proximityColor(model.Distance)
		drawText3x5(set, csX, 17, distanceText, distanceColor)

		dir := strings.ToUpper(strings.TrimSpace(model.Direction))
		if len(dir) > 2 {
			dir = dir[:2]
		}
		if dir != "" && dir != "--" {
			dirX := FrameWidth - textWidth3x5(dir)
			drawDirectionArrow(set, dirX-7, 17, dir, distanceColor)
			drawText3x5(set, dirX, 17, dir, distanceColor)
		}

		// Bottom telemetry uses one compact font per field. This makes the callsign
		// the only large text on screen and gives altitude/speed room to breathe.
		altText := compactAltitude(model.Altitude)
		altX := (32 - textWidth3x5(altText)) / 2
		if altX < 0 {
			altX = 0
		}
		drawText3x5(set, altX, 26, altText, altitudeColor(model.Altitude))

		speedText := compactSpeed(model.Speed)
		speedX := 32 + (32-textWidth3x5(speedText))/2
		if speedX < 32 {
			speedX = 32
		}
		drawText3x5(set, speedX, 26, speedText, speedColor(model.Speed))

	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawOperatorIcon(dst *image.RGBA, code string, x, y, size int) error {
	logo, err := assetpkg.OperatorIcon(code)
	if err != nil {
		return err
	}
	icon, err := png.Decode(bytes.NewReader(logo))
	if err != nil {
		return err
	}
	b := icon.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 || size <= 0 {
		return nil
	}
	for dy := 0; dy < size; dy++ {
		sy := b.Min.Y + dy*b.Dy()/size
		for dx := 0; dx < size; dx++ {
			sx := b.Min.X + dx*b.Dx()/size
			c := icon.At(sx, sy)
			_, _, _, a := c.RGBA()
			if a == 0 {
				continue
			}
			px, py := x+dx, y+dy
			if px >= 0 && px < FrameWidth && py >= 0 && py < FrameHeight {
				dst.Set(px, py, c)
			}
		}
	}
	return nil
}

func drawCentered5x7(setPixel func(int, int, color.Color), y int, text string, c color.Color) {
	text = strings.ToUpper(strings.TrimSpace(text))
	if text == "" {
		return
	}
	x := (FrameWidth - textWidth5x7(text)) / 2
	drawText5x7(setPixel, x, y, text, c)
}

func shortOperatorLabel(model DisplayModel) string {
	switch model.OperatorKind {
	case OperatorUnknown:
		return ""
	case OperatorMilitary:
		return trimLabel(model.OperatorCode, 6)
	case OperatorAirline:
		name := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(model.OperatorName), " ", ""))
		if name != "" && len(name) <= 6 {
			return name
		}
		return trimLabel(model.OperatorCode, 6)
	default:
		return trimLabel(model.OperatorCode, 6)
	}
}

func trimLabel(s string, max int) string {
	s = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), " ", ""))
	if len(s) > max {
		s = s[:max]
	}
	return s
}

func callsignColor(model DisplayModel) color.RGBA {
	code := strings.ToUpper(strings.TrimSpace(model.OperatorCode))
	if c, ok := carrierPalette[code]; ok {
		return c
	}

	// Some operators may not yet have a bundled logo/registry entry. Let a
	// recognizable ICAO callsign prefix still carry its identity color.
	cs := strings.ToUpper(strings.TrimSpace(model.Callsign))
	if len(cs) >= 3 {
		if c, ok := carrierPalette[cs[:3]]; ok {
			return c
		}
	}

	if model.OperatorKind == OperatorMilitary || model.Military {
		return green
	}
	if model.OperatorKind == OperatorUnknown {
		return cyan
	}
	return white
}

func proximityColor(s string) color.RGBA {
	nm, ok := parseTelemetryNumber(s, "NM")
	if !ok {
		return cyan
	}
	switch {
	case nm < 1:
		return magenta // OVERHEAD
	case nm < 2:
		return orange // VERY CLOSE
	case nm < 5:
		return amber // CLOSE
	case nm < 10:
		return green // NEARBY
	default:
		return cyan // TRACKED
	}
}

func altitudeColor(s string) color.RGBA {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "GND" {
		return magenta
	}
	feet, ok := parseTelemetryNumber(s, "FT")
	if !ok {
		return white
	}
	switch {
	case feet < 1000:
		return orange
	case feet < 5000:
		return amber
	case feet < 10000:
		return cyan
	default:
		return violet
	}
}

func speedColor(s string) color.RGBA {
	kt, ok := parseTelemetryNumber(s, "KT")
	if !ok {
		return white
	}
	switch {
	case kt < 100:
		return white
	case kt < 250:
		return green
	case kt < 400:
		return cyan
	default:
		return violet
	}
}

func parseTelemetryNumber(s, unit string) (float64, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSuffix(s, strings.ToUpper(unit))
	if s == "" || s == "--" || s == "---" || s == "GND" {
		return 0, false
	}
	if strings.HasSuffix(s, "K") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "K"), 64)
		return v * 1000, err == nil
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func compactDistanceValue(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimSpace(strings.TrimSuffix(s, "NM"))
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		if v >= 100 {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%.1f", v)
	}
	if len(s) > 5 {
		s = s[:5]
	}
	if s == "" {
		return "--"
	}
	return s
}

func compactAltitudeParts(s string) (string, string) {
	full := compactAltitude(s)
	if full == "GND" {
		return "GND", ""
	}
	if strings.HasSuffix(full, " FT") {
		return strings.TrimSuffix(full, " FT"), "FT"
	}
	return full, ""
}

func compactSpeedParts(s string) (string, string) {
	full := compactSpeed(s)
	if strings.HasSuffix(full, " KT") {
		return strings.TrimSuffix(full, " KT"), "KT"
	}
	return full, ""
}

func compactAltitude(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "GND" {
		return "GND"
	}
	if s == "" || s == "---" {
		return "-- FT"
	}
	n := strings.TrimSpace(strings.TrimSuffix(s, "FT"))
	feet, err := strconv.Atoi(n)
	if err != nil {
		return trimLabel(n, 5) + " FT"
	}
	if feet >= 10000 {
		return fmt.Sprintf("%.0fK FT", float64(feet)/1000)
	}
	if feet >= 1000 {
		return fmt.Sprintf("%.1fK FT", float64(feet)/1000)
	}
	return fmt.Sprintf("%d FT", feet)
}

func compactSpeed(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" || s == "---" {
		return "-- KT"
	}
	n := strings.TrimSpace(strings.TrimSuffix(s, "KT"))
	if len(n) > 3 {
		n = n[:3]
	}
	return n + " KT"
}

func directionVector(direction string) (int, int, bool) {
	direction = strings.ToUpper(strings.TrimSpace(direction))
	vectors := map[string][2]int{
		"N":  {0, -1},
		"NE": {1, -1},
		"E":  {1, 0},
		"SE": {1, 1},
		"S":  {0, 1},
		"SW": {-1, 1},
		"W":  {-1, 0},
		"NW": {-1, -1},
	}
	v, ok := vectors[direction]
	if !ok {
		return 0, 0, false
	}
	return v[0], v[1], true
}

func drawDirectionArrow(setPixel func(int, int, color.Color), x, y int, direction string, c color.Color) {
	vx, vy, ok := directionVector(direction)
	if !ok {
		return
	}
	cx, cy := x+2, y+2
	for i := -2; i <= 2; i++ {
		setPixel(cx+i*vx, cy+i*vy, c)
	}
	tipX, tipY := cx+2*vx, cy+2*vy
	baseX, baseY := tipX-vx, tipY-vy
	px, py := -vy, vx
	setPixel(baseX+px, baseY+py, c)
	setPixel(baseX-px, baseY-py, c)
}

func idleLines(name string) (string, string) {
	name = strings.ToUpper(strings.TrimSpace(name))
	if name == "" {
		name = "ADS-B MATRIX"
	}
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		line1 := parts[0]
		line2 := strings.Join(parts[1:], "")
		if len(line1) > 8 {
			line1 = line1[:8]
		}
		if len(line2) > 8 {
			line2 = line2[:8]
		}
		return line1, line2
	}
	if len(name) <= 8 {
		return name, "TRACKING"
	}
	return name[:8], name[8:minInt(len(name), 16)]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
