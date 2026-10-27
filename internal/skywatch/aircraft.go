package skywatch

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

const earthRadiusNM = 3440.065

type Dump1090Response struct {
	Now      float64       `json:"now"`
	Messages int64         `json:"messages"`
	Aircraft []AircraftRaw `json:"aircraft"`
}

type AircraftRaw struct {
	Hex          string          `json:"hex"`
	Flight       string          `json:"flight"`
	Type         string          `json:"t"`
	Lat          *float64        `json:"lat"`
	Lon          *float64        `json:"lon"`
	Seen         float64         `json:"seen"`
	SeenPos      *float64        `json:"seen_pos"`
	AltBaro      json.RawMessage `json:"alt_baro"`
	AltGeom      json.RawMessage `json:"alt_geom"`
	GS           *float64        `json:"gs"`
	Track        *float64        `json:"track"`
	Squawk       string          `json:"squawk"`
	Category     string          `json:"category"`
	RSSI         *float64        `json:"rssi"`
	Registration string          `json:"r"`
	DBFlags      int             `json:"dbFlags"`
	OwnOp        string          `json:"ownOp"`
	Description  string          `json:"desc"`
}

type Aircraft struct {
	Hex            string       `json:"hex"`
	Callsign       string       `json:"callsign"`
	OperatorCode   string       `json:"operator_code"`
	OperatorName   string       `json:"operator_name"`
	OperatorKind   OperatorKind `json:"operator_kind"`
	Military       bool         `json:"military"`
	MilitaryBranch string       `json:"military_branch,omitempty"`
	Detection      string       `json:"operator_detection,omitempty"`
	LogoURL        string       `json:"logo_url"`

	// Legacy aliases retained for clients built against the first API version.
	OperatorICAO  string   `json:"operator_icao"`
	AirlineName   string   `json:"airline_name"`
	Type          string   `json:"type,omitempty"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	DistanceNM    float64  `json:"distance_nm"`
	BearingDeg    float64  `json:"bearing_deg"`
	Direction     string   `json:"direction"`
	AltitudeFT    *int     `json:"altitude_ft,omitempty"`
	OnGround      bool     `json:"on_ground"`
	GroundSpeed   *float64 `json:"ground_speed_kt,omitempty"`
	TrackDeg      *float64 `json:"track_deg,omitempty"`
	SeenSeconds   float64  `json:"seen_seconds"`
	SeenPosSecs   float64  `json:"seen_position_seconds"`
	Squawk        string   `json:"squawk,omitempty"`
	Category      string   `json:"category,omitempty"`
	SignalDBFS    *float64 `json:"signal_dbfs,omitempty"`
	Registration  string   `json:"registration,omitempty"`
	OwnerOperator string   `json:"owner_operator,omitempty"`
	DBFlags       int      `json:"db_flags,omitempty"`
}

type Filter struct {
	ReceiverLat    float64
	ReceiverLon    float64
	MaxPositionAge time.Duration
	MaxDistanceNM  float64
	MinAltitudeFT  int
	IncludeGround  bool
}

func Normalize(raw AircraftRaw, f Filter) (Aircraft, bool) {
	if raw.Lat == nil || raw.Lon == nil {
		return Aircraft{}, false
	}
	if *raw.Lat < -90 || *raw.Lat > 90 || *raw.Lon < -180 || *raw.Lon > 180 {
		return Aircraft{}, false
	}

	seenPos := raw.Seen
	if raw.SeenPos != nil {
		seenPos = *raw.SeenPos
	}
	if f.MaxPositionAge > 0 && seenPos > f.MaxPositionAge.Seconds() {
		return Aircraft{}, false
	}

	distance := HaversineNM(f.ReceiverLat, f.ReceiverLon, *raw.Lat, *raw.Lon)
	if f.MaxDistanceNM > 0 && distance > f.MaxDistanceNM {
		return Aircraft{}, false
	}

	altitude, onGround := parseAltitude(raw.AltBaro)
	if altitude == nil && !onGround {
		altitude, onGround = parseAltitude(raw.AltGeom)
	}
	if onGround && !f.IncludeGround {
		return Aircraft{}, false
	}
	if altitude != nil && *altitude < f.MinAltitudeFT {
		return Aircraft{}, false
	}

	callsign := strings.TrimSpace(raw.Flight)
	if callsign == "" {
		callsign = strings.ToUpper(strings.TrimSpace(raw.Hex))
	}

	bearing := InitialBearing(f.ReceiverLat, f.ReceiverLon, *raw.Lat, *raw.Lon)
	op := ResolveOperator(raw, callsign)

	return Aircraft{
		Hex:            strings.ToUpper(strings.TrimSpace(raw.Hex)),
		Callsign:       callsign,
		OperatorCode:   op.Code,
		OperatorName:   op.Name,
		OperatorKind:   op.Kind,
		Military:       op.IsMilitary,
		MilitaryBranch: op.Branch,
		Detection:      op.Detection,
		OperatorICAO:   op.Code,
		AirlineName:    op.Name,
		LogoURL:        op.LogoURL,
		Type:           strings.ToUpper(strings.TrimSpace(raw.Type)),
		Latitude:       *raw.Lat,
		Longitude:      *raw.Lon,
		DistanceNM:     distance,
		BearingDeg:     bearing,
		Direction:      Cardinal8(bearing),
		AltitudeFT:     altitude,
		OnGround:       onGround,
		GroundSpeed:    raw.GS,
		TrackDeg:       raw.Track,
		SeenSeconds:    raw.Seen,
		SeenPosSecs:    seenPos,
		Squawk:         raw.Squawk,
		Category:       raw.Category,
		SignalDBFS:     raw.RSSI,
		Registration:   strings.ToUpper(strings.TrimSpace(raw.Registration)),
		OwnerOperator:  strings.TrimSpace(raw.OwnOp),
		DBFlags:        raw.DBFlags,
	}, true
}

func Nearest(raw []AircraftRaw, f Filter) (*Aircraft, []Aircraft) {
	normalized := make([]Aircraft, 0, len(raw))
	for _, item := range raw {
		a, ok := Normalize(item, f)
		if !ok {
			continue
		}
		normalized = append(normalized, a)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].DistanceNM < normalized[j].DistanceNM
	})
	if len(normalized) == 0 {
		return nil, normalized
	}
	nearest := normalized[0]
	return &nearest, normalized
}

func HaversineNM(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := math.Pi / 180
	p1 := lat1 * toRad
	p2 := lat2 * toRad
	dp := (lat2 - lat1) * toRad
	dl := (lon2 - lon1) * toRad

	a := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusNM * c
}

func InitialBearing(lat1, lon1, lat2, lon2 float64) float64 {
	toRad := math.Pi / 180
	p1 := lat1 * toRad
	p2 := lat2 * toRad
	dl := (lon2 - lon1) * toRad
	y := math.Sin(dl) * math.Cos(p2)
	x := math.Cos(p1)*math.Sin(p2) - math.Sin(p1)*math.Cos(p2)*math.Cos(dl)
	deg := math.Atan2(y, x) * 180 / math.Pi
	return math.Mod(deg+360, 360)
}

func Cardinal8(deg float64) string {
	dirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int(math.Floor((deg+22.5)/45.0)) % 8
	return dirs[idx]
}

func parseAltitude(raw json.RawMessage) (*int, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, false
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if strings.EqualFold(s, "ground") {
			return nil, true
		}
		if n, err := strconv.ParseFloat(s, 64); err == nil {
			v := int(math.Round(n))
			return &v, false
		}
		return nil, false
	}

	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		v := int(math.Round(n))
		return &v, false
	}
	return nil, false
}

func FormatDistance(nm float64) string {
	if nm < 10 {
		return fmt.Sprintf("%.1fNM", nm)
	}
	return fmt.Sprintf("%.0fNM", nm)
}
