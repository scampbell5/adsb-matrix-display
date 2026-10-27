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
