package skywatch

import "testing"

func TestResolveAirlineCompatibility(t *testing.T) {
	got := ResolveAirline("UAL1827")
	if got.Code != "UAL" || got.Name != "United" || got.Kind != OperatorAirline {
		t.Fatalf("unexpected airline: %+v", got)
	}
	unknown := ResolveAirline("N12345")
	if unknown.Code != "UNK" || unknown.Kind != OperatorUnknown {
		t.Fatalf("expected UNK, got %+v", unknown)
	}
}

func TestResolveMilitaryByDBFlags(t *testing.T) {
	raw := AircraftRaw{Hex: "AE1234", DBFlags: 1, OwnOp: "UNITED STATES AIR FORCE"}
	got := ResolveOperator(raw, "RCH421")
	if got.Code != "USAF" || !got.IsMilitary || got.Kind != OperatorMilitary {
		t.Fatalf("unexpected military operator: %+v", got)
	}
}

func TestResolveMilitaryByUSHexRange(t *testing.T) {
	raw := AircraftRaw{Hex: "AE1234"}
	got := ResolveOperator(raw, "UNKNOWN")
	if got.Code != "MIL" || !got.IsMilitary {
		t.Fatalf("expected generic military, got %+v", got)
	}
}

func TestResolveMilitaryBranchFromCallsign(t *testing.T) {
	got := ResolveOperator(AircraftRaw{Hex: "AE1234"}, "RCH421")
	if got.Code != "USAF" {
		t.Fatalf("expected USAF, got %+v", got)
	}
	army := ResolveOperator(AircraftRaw{Hex: "AE2345"}, "PAT123")
	if army.Code != "USA" {
		t.Fatalf("expected Army, got %+v", army)
	}
}
