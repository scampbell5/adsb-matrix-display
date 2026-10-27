package webui

import _ "embed"

// DisplayHTML is the built-in browser preview for the 64x32 hardware frame.
//
//go:embed display.html
var DisplayHTML []byte
