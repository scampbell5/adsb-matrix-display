package skywatch

import (
	"fmt"
	"math"
	"strings"
)

type DisplayModel struct {
	DisplayName    string       `json:"display_name"`
	Callsign       string       `json:"callsign"`
	Type           string       `json:"type"`
	OperatorCode   string       `json:"operator_code"`
	OperatorName   string       `json:"operator_name"`
	OperatorKind   OperatorKind `json:"operator_kind"`
	Military       bool         `json:"military"`
	MilitaryBranch string       `json:"military_branch,omitempty"`
	LogoURL        string       `json:"logo_url"`

	// Legacy aliases retained for compatibility.
	OperatorICAO string `json:"operator_icao"`
	AirlineName  string `json:"airline_name"`

	Distance    string `json:"distance"`
	Direction   string `json:"direction"`
	Altitude    string `json:"altitude"`
	Speed       string `json:"speed"`
	Idle        bool   `json:"idle"`
	Fingerprint string `json:"fingerprint"`
}

func BuildDisplay(a *Aircraft, displayName string) DisplayModel {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = "ADS-B Matrix"
	}
	if a == nil {
		return DisplayModel{
			DisplayName: displayName, Callsign: displayName, OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: OperatorUnknown,
			OperatorICAO: "UNK", AirlineName: "Unknown", LogoURL: "/assets/operators/UNK.png",
			Distance: "NO POSITION", Idle: true, Fingerprint: "idle",
		}
	}

	callsign := strings.ToUpper(strings.TrimSpace(a.Callsign))
	if callsign == "" {
		callsign = a.Hex
	}
	if len(callsign) > 8 {
		callsign = callsign[:8]
	}

	typ := strings.ToUpper(strings.TrimSpace(a.Type))
	if len(typ) > 4 {
		typ = typ[:4]
	}

	altitude := "---"
	altitudeBucket := -1
	if a.OnGround {
		altitude = "GND"
		altitudeBucket = 0
	} else if a.AltitudeFT != nil {
		rounded := int(math.Round(float64(*a.AltitudeFT)/100.0) * 100)
		altitude = fmt.Sprintf("%dFT", rounded)
		altitudeBucket = rounded
	}

	speed := "---"
	speedBucket := -1
	if a.GroundSpeed != nil {
		rounded := int(math.Round(*a.GroundSpeed/5.0) * 5)
		speed = fmt.Sprintf("%dKT", rounded)
		speedBucket = rounded
	}

	distance := FormatDistance(a.DistanceNM)
	distanceBucket := math.Round(a.DistanceNM*10) / 10
	fp := fmt.Sprintf("%s|%s|%s|%.1f|%s|%d|%d", callsign, typ, a.OperatorCode, distanceBucket, a.Direction, altitudeBucket, speedBucket)

	return DisplayModel{
		DisplayName: displayName, Callsign: callsign, Type: typ,
		OperatorCode: a.OperatorCode, OperatorName: a.OperatorName, OperatorKind: a.OperatorKind,
		Military: a.Military, MilitaryBranch: a.MilitaryBranch, LogoURL: a.LogoURL,
		OperatorICAO: a.OperatorCode, AirlineName: a.OperatorName,
		Distance: distance, Direction: a.Direction, Altitude: altitude, Speed: speed, Fingerprint: fp,
	}
}
