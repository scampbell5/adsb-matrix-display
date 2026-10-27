package skywatch

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestHaversineNM(t *testing.T) {
	// One degree of longitude at the equator is about 60 nautical miles.
	got := HaversineNM(0, 0, 0, 1)
	if math.Abs(got-60) > 1 {
		t.Fatalf("unexpected distance: %.1f NM", got)
	}
}

func TestCardinal8(t *testing.T) {
	cases := map[float64]string{0: "N", 44: "NE", 90: "E", 181: "S", 270: "W", 315: "NW"}
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
		{Hex: "abc001", Flight: "FRESH1", Lat: ptr(0.10), Lon: ptr(0.10), SeenPos: &fresh, AltBaro: alt},
		{Hex: "abc002", Flight: "STALE1", Lat: ptr(0.05), Lon: ptr(0.05), SeenPos: &stale, AltBaro: alt},
	}
	nearest, all := Nearest(raw, Filter{ReceiverLat: 0, ReceiverLon: 0, MaxPositionAge: 10 * time.Second})
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
	d := BuildDisplay(&Aircraft{Callsign: "UAL1827", Type: "B738", DistanceNM: 2.84, Direction: "SW", AltitudeFT: &alt, GroundSpeed: &speed}, "ADS-B Matrix")
	if d.Distance != "2.8NM" || d.Altitude != "6400FT" || d.Speed != "220KT" {
		t.Fatalf("unexpected display: %#v", d)
	}
}

func ptr(v float64) *float64 { return &v }
