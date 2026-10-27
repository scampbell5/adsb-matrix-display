package assets

import (
	"embed"
	"fmt"
	"path"
	"strings"
)

// Operator icons are original, pixel-friendly identifier marks made for this
// project. Every icon is exactly 20x20 RGBA and is intended for the 64x32
// matrix renderer. The renderer currently displays them at 15x15 using
// nearest-neighbor sampling so the logo remains crisp while leaving more room for data.
//
//go:embed operators/*.png
var operatorFS embed.FS

func OperatorIcon(code string) ([]byte, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" || code == "UNKNOWN" {
		code = "UNK"
	}
	name := path.Join("operators", code+".png")
	b, err := operatorFS.ReadFile(name)
	if err != nil && code != "UNK" {
		return OperatorIcon("UNK")
	}
	if err != nil {
		return nil, fmt.Errorf("operator icon %s: %w", code, err)
	}
	return b, nil
}

// AirlineLogo is kept for API/backwards compatibility with the first project
// version. New code should use OperatorIcon.
func AirlineLogo(code string) ([]byte, error) { return OperatorIcon(code) }
