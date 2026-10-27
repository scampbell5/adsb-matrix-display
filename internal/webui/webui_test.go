package webui

import (
	"bytes"
	"testing"
)

func TestDisplayHTMLIsEmbedded(t *testing.T) {
	if len(DisplayHTML) < 1000 {
		t.Fatalf("display HTML unexpectedly small: %d bytes", len(DisplayHTML))
	}
	for _, want := range [][]byte{
		[]byte("ADS-B Matrix"),
		[]byte("64 × 32"),
		[]byte("/api/v1/display/frame.png"),
		[]byte("/api/v1/status"),
	} {
		if !bytes.Contains(DisplayHTML, want) {
			t.Fatalf("display HTML missing %q", want)
		}
	}
}
