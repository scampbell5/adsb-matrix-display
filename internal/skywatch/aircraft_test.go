package skywatch

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestHaversineNM(t *testing.T) {
	// Approximate great-circle distance BOS -> JFK is about 162 NM.
	got := HaversineNM(42.3656, -71.0096, 40.6413, -73.7781)
	if math.Abs(got-162) > 8 {
		t.Fatalf("unexpected distance: %.1f NM", got)
	}
}

func TestCardinal8(t *testing.T) {
	cases := map[float64]string{0: "N", 45: "NE", 90: "E", 135: "SE", 180: "S", 225: "SW", 270: "W", 315: "NW", 359: "N"}
	for deg, want := range cases {
		if got := Cardinal8(deg); got != want {
			t.Fatalf("Cardinal8(%v)=%q want %q", deg, got, want)
		}
	}
}

func TestNearestFiltersStalePositions(t *testing.T) {
	alt, _ := json.Marshal(5200)
	fresh := 1.0
	stale := 30.0
	raw := []AircraftRaw{
		{Hex: "abc001", Flight: "FRESH1", Lat: ptr(42.4), Lon: ptr(-71.1), SeenPos: &fresh, AltBaro: alt},
		{Hex: "abc002", Flight: "STALE1", Lat: ptr(42.1), Lon: ptr(-71.1), SeenPos: &stale, AltBaro: alt},
	}
	nearest, all := Nearest(raw, Filter{ReceiverLat: 42.3, ReceiverLon: -71.0, MaxPositionAge: 10 * time.Second})
	if nearest == nil || nearest.Callsign != "FRESH1" {
		t.Fatalf("unexpected nearest: %#v", nearest)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 normalized aircraft, got %d", len(all))
	}
}

func TestBuildDisplayBucketsData(t *testing.T) {
	alt := 6438
	speed := 222.1
	d := BuildDisplay(&Aircraft{Callsign: "UAL1827", Type: "B738", DistanceNM: 2.84, Direction: "SW", AltitudeFT: &alt, GroundSpeed: &speed})
	if d.Distance != "2.8NM" || d.Altitude != "6400FT" || d.Speed != "220KT" {
		t.Fatalf("unexpected display: %#v", d)
	}
}

func ptr(v float64) *float64 { return &v }
