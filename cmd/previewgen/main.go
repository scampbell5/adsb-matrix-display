package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"adsb-matrix-display/internal/skywatch"
)

func main() {
	out := "examples/frames"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		log.Fatal(err)
	}

	models := map[string]skywatch.DisplayModel{
		"united.png":    {Callsign: "UAL1827", Type: "B738", OperatorCode: "UAL", OperatorName: "United", OperatorKind: skywatch.OperatorAirline, LogoURL: "/assets/operators/UAL.png", Distance: "2.8NM", Direction: "SW", Altitude: "6400FT", Speed: "220KT"},
		"delta.png":     {Callsign: "DAL2416", Type: "A321", OperatorCode: "DAL", OperatorName: "Delta", OperatorKind: skywatch.OperatorAirline, LogoURL: "/assets/operators/DAL.png", Distance: "5.4NM", Direction: "NE", Altitude: "8300FT", Speed: "245KT"},
		"southwest.png": {Callsign: "SWA593", Type: "B737", OperatorCode: "SWA", OperatorName: "Southwest", OperatorKind: skywatch.OperatorAirline, LogoURL: "/assets/operators/SWA.png", Distance: "1.7NM", Direction: "W", Altitude: "3100FT", Speed: "185KT"},
		"usaf.png":      {Callsign: "RCH421", Type: "C17", OperatorCode: "USAF", OperatorName: "United States Air Force", OperatorKind: skywatch.OperatorMilitary, Military: true, MilitaryBranch: "Air Force", LogoURL: "/assets/operators/USAF.png", Distance: "4.2NM", Direction: "NE", Altitude: "8200FT", Speed: "250KT"},
		"military.png":  {Callsign: "AE1234", Type: "H60", OperatorCode: "MIL", OperatorName: "Military Aircraft", OperatorKind: skywatch.OperatorMilitary, Military: true, MilitaryBranch: "Unknown", LogoURL: "/assets/operators/MIL.png", Distance: "7.1NM", Direction: "S", Altitude: "2400FT", Speed: "135KT"},
		"unknown.png":   {Callsign: "N1410A", Type: "", OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: skywatch.OperatorUnknown, LogoURL: "/assets/operators/UNK.png", Distance: "14.0NM", Direction: "NE", Altitude: "400FT", Speed: "85KT"},
		"n380f.png":     {Callsign: "N380F", Type: "", OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: skywatch.OperatorUnknown, LogoURL: "/assets/operators/UNK.png", Distance: "7.7NM", Direction: "SE", Altitude: "6100FT", Speed: "155KT"},
		"capeair.png":   {Callsign: "KAP2112", Type: "C402", OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: skywatch.OperatorUnknown, LogoURL: "/assets/operators/UNK.png", Distance: "6.8NM", Direction: "S", Altitude: "5600FT", Speed: "145KT"},
		"overhead.png":  {Callsign: "N603NH", Type: "", OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: skywatch.OperatorUnknown, LogoURL: "/assets/operators/UNK.png", Distance: "0.7NM", Direction: "NW", Altitude: "800FT", Speed: "90KT"},
		"idle.png":      {Callsign: "ADS-B MATRIX", OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: skywatch.OperatorUnknown, LogoURL: "/assets/operators/UNK.png", Distance: "NO POSITION", Idle: true, Fingerprint: "idle"},
	}

	// Compass coverage samples: all eight cardinal/intercardinal directions use
	// the same proximity color as distance so the navigation row reads as one signal.
	for _, dir := range []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"} {
		models["direction_"+strings.ToLower(dir)+".png"] = skywatch.DisplayModel{
			Callsign: "N603NH", OperatorCode: "UNK", OperatorName: "Unknown", OperatorKind: skywatch.OperatorUnknown,
			LogoURL: "/assets/operators/UNK.png", Distance: "6.8NM", Direction: dir, Altitude: "5600FT", Speed: "145KT",
		}
	}

	for name, model := range models {
		b, err := skywatch.RenderFramePNG(model)
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, name), b, 0o644); err != nil {
			log.Fatal(err)
		}
	}
}
