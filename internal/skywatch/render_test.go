package skywatch

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderFramePNG(t *testing.T) {
	model := DisplayModel{
		Callsign: "UAL1827", Type: "B738", OperatorCode: "UAL", OperatorName: "United", OperatorKind: OperatorAirline,
		OperatorICAO: "UAL", AirlineName: "United", LogoURL: "/assets/operators/UAL.png", Distance: "2.8NM", Direction: "SW",
		Altitude: "6400FT", Speed: "220KT", Fingerprint: "test",
	}
	b, err := RenderFramePNG(model)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != FrameWidth || img.Bounds().Dy() != FrameHeight {
		t.Fatalf("unexpected size: %v", img.Bounds())
	}
}

func TestRenderMilitaryFramePNG(t *testing.T) {
	model := DisplayModel{
		Callsign: "RCH421", Type: "C17", OperatorCode: "USAF", OperatorName: "United States Air Force", OperatorKind: OperatorMilitary,
		Military: true, MilitaryBranch: "Air Force", LogoURL: "/assets/operators/USAF.png", Distance: "4.2NM", Direction: "NE",
		Altitude: "8200FT", Speed: "250KT", Fingerprint: "mil-test",
	}
	b, err := RenderFramePNG(model)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 32 {
		t.Fatalf("unexpected size: %v", img.Bounds())
	}
}

func TestCompactTelemetryFormatting(t *testing.T) {
	if got := compactAltitude("6100FT"); got != "6.1K FT" {
		t.Fatalf("compactAltitude = %q, want %q", got, "6.1K FT")
	}
	if got := compactSpeed("155KT"); got != "155 KT" {
		t.Fatalf("compactSpeed = %q, want %q", got, "155 KT")
	}
	if got := compactDistanceValue("7.7NM"); got != "7.7" {
		t.Fatalf("compactDistanceValue = %q, want %q", got, "7.7")
	}
}

func TestEightCharacterCallsignRenders(t *testing.T) {
	model := DisplayModel{
		Callsign: "ABCDEFGH", Type: "B738", OperatorCode: "UAL", OperatorName: "United", OperatorKind: OperatorAirline,
		Distance: "9.9NM", Direction: "SE", Altitude: "6100FT", Speed: "155KT",
	}
	b, err := RenderFramePNG(model)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 32 {
		t.Fatalf("unexpected size: %v", img.Bounds())
	}
}

func TestGamifiedTelemetryColors(t *testing.T) {
	if got := proximityColor("14.0NM"); got != cyan {
		t.Fatalf("far proximity color = %#v, want cyan", got)
	}
	if got := proximityColor("6.8NM"); got != green {
		t.Fatalf("nearby proximity color = %#v, want green", got)
	}
	if got := proximityColor("3.2NM"); got != amber {
		t.Fatalf("close proximity color = %#v, want amber", got)
	}
	if got := proximityColor("1.4NM"); got != orange {
		t.Fatalf("very-close proximity color = %#v, want orange", got)
	}
	if got := proximityColor("0.7NM"); got != magenta {
		t.Fatalf("overhead proximity color = %#v, want magenta", got)
	}

	if got := altitudeColor("400FT"); got != orange {
		t.Fatalf("low altitude color = %#v, want orange", got)
	}
	if got := altitudeColor("3900FT"); got != amber {
		t.Fatalf("mid-low altitude color = %#v, want amber", got)
	}
	if got := altitudeColor("5600FT"); got != cyan {
		t.Fatalf("mid-high altitude color = %#v, want cyan", got)
	}
	if got := altitudeColor("12000FT"); got != violet {
		t.Fatalf("high altitude color = %#v, want violet", got)
	}

	if got := speedColor("85KT"); got != white {
		t.Fatalf("slow speed color = %#v, want white", got)
	}
	if got := speedColor("145KT"); got != green {
		t.Fatalf("cruise speed color = %#v, want green", got)
	}
	if got := speedColor("300KT"); got != cyan {
		t.Fatalf("fast speed color = %#v, want cyan", got)
	}
	if got := speedColor("430KT"); got != violet {
		t.Fatalf("very-fast speed color = %#v, want violet", got)
	}
}

func TestCallsignCarrierColors(t *testing.T) {
	delta := DisplayModel{Callsign: "DAL2416", OperatorCode: "DAL", OperatorKind: OperatorAirline}
	if got := callsignColor(delta); got != carrierPalette["DAL"] {
		t.Fatalf("Delta callsign color = %#v, want %#v", got, carrierPalette["DAL"])
	}

	// KAP does not need a bundled logo to receive its brand-inspired callsign color.
	capeAir := DisplayModel{Callsign: "KAP2112", OperatorCode: "UNK", OperatorKind: OperatorUnknown}
	if got := callsignColor(capeAir); got != carrierPalette["KAP"] {
		t.Fatalf("KAP prefix callsign color = %#v, want %#v", got, carrierPalette["KAP"])
	}

	ga := DisplayModel{Callsign: "N380F", OperatorCode: "UNK", OperatorKind: OperatorUnknown}
	if got := callsignColor(ga); got != cyan {
		t.Fatalf("GA callsign color = %#v, want cyan", got)
	}
}
