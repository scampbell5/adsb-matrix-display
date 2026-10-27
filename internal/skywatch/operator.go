package skywatch

import (
	"strconv"
	"strings"
)

type OperatorKind string

const (
	OperatorAirline  OperatorKind = "airline"
	OperatorMilitary OperatorKind = "military"
	OperatorUnknown  OperatorKind = "unknown"
)

type OperatorInfo struct {
	Code       string       `json:"code"`
	Name       string       `json:"name"`
	Kind       OperatorKind `json:"kind"`
	Branch     string       `json:"branch,omitempty"`
	LogoURL    string       `json:"logo_url"`
	Detection  string       `json:"detection,omitempty"`
	IsMilitary bool         `json:"is_military"`
}

var airlineRegistry = map[string]OperatorInfo{
	"DAL": airline("DAL", "Delta"),
	"UAL": airline("UAL", "United"),
	"AAL": airline("AAL", "American"),
	"SWA": airline("SWA", "Southwest"),
	"JBU": airline("JBU", "JetBlue"),
	"ASA": airline("ASA", "Alaska"),
	"NKS": airline("NKS", "Spirit"),
	"FFT": airline("FFT", "Frontier"),
	"AAY": airline("AAY", "Allegiant"),
	"RPA": airline("RPA", "Republic Airways"),
	"EDV": airline("EDV", "Endeavor Air"),
	"ENY": airline("ENY", "Envoy Air"),
	"JIA": airlineWithIcon("JIA", "PSA Airlines", "JIA"),
	"SKW": airline("SKW", "SkyWest"),
	"AWI": airline("AWI", "Air Wisconsin"),
	"ACA": airline("ACA", "Air Canada"),
}

var militaryRegistry = map[string]OperatorInfo{
	"USAF": military("USAF", "United States Air Force", "Air Force"),
	"USMC": military("USMC", "United States Marine Corps", "Marine Corps"),
	"USCG": military("USCG", "United States Coast Guard", "Coast Guard"),
	"USN":  military("USN", "United States Navy", "Navy"),
	"USA":  military("USA", "United States Army", "Army"),
	"MIL":  military("MIL", "Military Aircraft", "Unknown"),
}

func airline(code, name string) OperatorInfo { return airlineWithIcon(code, name, code) }
func airlineWithIcon(code, name, icon string) OperatorInfo {
	return OperatorInfo{Code: code, Name: name, Kind: OperatorAirline, LogoURL: "/assets/operators/" + icon + ".png"}
}
func military(code, name, branch string) OperatorInfo {
	return OperatorInfo{Code: code, Name: name, Kind: OperatorMilitary, Branch: branch, LogoURL: "/assets/operators/" + code + ".png", IsMilitary: true}
}
func unknownOperator() OperatorInfo {
	return OperatorInfo{Code: "UNK", Name: "Unknown", Kind: OperatorUnknown, LogoURL: "/assets/operators/UNK.png"}
}

// ResolveOperator deliberately checks military status before airline callsign
// prefixes. A military aircraft can transmit a callsign that looks unrelated
// to its service branch, while readsb dbFlags / owner metadata / known US
// military ICAO ranges are stronger signals.
func ResolveOperator(raw AircraftRaw, callsign string) OperatorInfo {
	owner := strings.ToUpper(strings.TrimSpace(raw.OwnOp))
	if branch := branchFromOwner(owner); branch != "" {
		op := militaryRegistry[branch]
		op.Detection = "owner"
		return op
	}

	militaryByDB := raw.DBFlags&1 != 0
	militaryByRange := isUSMilitaryICAO(raw.Hex)
	if militaryByDB || militaryByRange {
		if branch := branchFromCallsign(callsign); branch != "" {
			op := militaryRegistry[branch]
			op.Detection = "military_flag+callsign"
			return op
		}
		op := militaryRegistry["MIL"]
		if militaryByDB {
			op.Detection = "dbFlags"
		} else {
			op.Detection = "us_military_icao_range"
		}
		return op
	}

	// A small set of highly characteristic military callsign prefixes is used
	// only as a fallback when richer database information is absent.
	if branch := branchFromCallsign(callsign); branch != "" {
		op := militaryRegistry[branch]
		op.Detection = "callsign_heuristic"
		return op
	}

	s := strings.ToUpper(strings.TrimSpace(callsign))
	if len(s) >= 3 {
		if a, ok := airlineRegistry[s[:3]]; ok {
			a.Detection = "callsign_prefix"
			return a
		}
	}
	return unknownOperator()
}

func branchFromOwner(owner string) string {
	switch {
	case strings.Contains(owner, "COAST GUARD") || strings.Contains(owner, "USCG"):
		return "USCG"
	case strings.Contains(owner, "MARINE CORPS") || strings.Contains(owner, "USMC"):
		return "USMC"
	case strings.Contains(owner, "UNITED STATES NAVY") || strings.Contains(owner, "US NAVY"):
		return "USN"
	case strings.Contains(owner, "UNITED STATES ARMY") || strings.Contains(owner, "US ARMY"):
		return "USA"
	case strings.Contains(owner, "AIR FORCE") || strings.Contains(owner, "USAF"):
		return "USAF"
	}
	return ""
}

func branchFromCallsign(callsign string) string {
	s := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(callsign), " ", ""))
	// Deliberately conservative. These are intended as a fallback, not a
	// complete military callsign database.
	for _, p := range []string{"RCH", "SAM", "SPAR", "VENUS"} {
		if strings.HasPrefix(s, p) {
			return "USAF"
		}
	}
	for _, p := range []string{"NAVY", "CNV"} {
		if strings.HasPrefix(s, p) {
			return "USN"
		}
	}
	if strings.HasPrefix(s, "PAT") {
		return "USA"
	}
	if strings.HasPrefix(s, "USCG") || strings.HasPrefix(s, "CGRD") {
		return "USCG"
	}
	if strings.HasPrefix(s, "USMC") {
		return "USMC"
	}
	return ""
}

func isUSMilitaryICAO(hex string) bool {
	s := strings.ToLower(strings.TrimSpace(hex))
	s = strings.TrimPrefix(s, "~")
	if len(s) != 6 {
		return false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return false
	}
	// Mirrors readsb's U.S. military range detection:
	// ADF7C8 through AFFFFF (including AE0000-AFFFFF).
	return n >= 0xADF7C8 && n <= 0xAFFFFF
}

func Operators() []OperatorInfo {
	order := []string{"DAL", "UAL", "AAL", "SWA", "JBU", "ASA", "NKS", "FFT", "AAY", "RPA", "EDV", "ENY", "JIA", "SKW", "AWI", "ACA"}
	out := make([]OperatorInfo, 0, len(order)+len(militaryRegistry)+1)
	for _, code := range order {
		out = append(out, airlineRegistry[code])
	}
	for _, code := range []string{"USAF", "USMC", "USCG", "USN", "USA", "MIL"} {
		out = append(out, militaryRegistry[code])
	}
	out = append(out, unknownOperator())
	return out
}

func Airlines() []OperatorInfo {
	order := []string{"DAL", "UAL", "AAL", "SWA", "JBU", "ASA", "NKS", "FFT", "AAY", "RPA", "EDV", "ENY", "JIA", "SKW", "AWI", "ACA"}
	out := make([]OperatorInfo, 0, len(order))
	for _, code := range order {
		out = append(out, airlineRegistry[code])
	}
	return out
}

// ResolveAirline remains for compatibility with older callers/tests.
func ResolveAirline(callsign string) OperatorInfo {
	s := strings.ToUpper(strings.TrimSpace(callsign))
	if len(s) >= 3 {
		if a, ok := airlineRegistry[s[:3]]; ok {
			return a
		}
	}
	return unknownOperator()
}
