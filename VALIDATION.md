# Validation

Final v8 carrier-color + gamified telemetry + explicit 8-way compass 64x32 display build.

- `go test ./internal/skywatch ./internal/webui ./internal/config` — PASS
- `go run ./cmd/previewgen` — PASS
- Generated frames are exactly 64x32 PNGs.
- Unknown/UFO artwork renders at 14x14; known operator artwork renders at 16x16.
- Callsign is the only large 5x7 text field; navigation and bottom telemetry use the compact 3x5 font.
- Callsign, proximity, altitude, and speed colors are data-driven with tested threshold bands.
- All eight directions (N, NE, E, SE, S, SW, W, NW) are renderer-tested and generated as sample frames.
- Cardinal and diagonal arrows share the same proximity color as the distance and direction label.
- Unknown/GA callsigns remain cyan; known carrier callsigns use LED-friendly carrier identity colors.
- Dockerfile uses Go 1.25 Alpine and builds with `go build -mod=mod`.
- Dockerfile contains no executable `RUN go mod download` step.
- Compose retains configurable host port, PiAware URL, and external `adsb-net` network.

The complete Docker image build still depends on resolving `github.com/alparslanahmed/huidu-led` from the Docker host during `go build -mod=mod`.

## Full-suite note

`go test ./...` / `go test -mod=mod ./...` cannot complete in this ChatGPT container because outbound DNS to `proxy.golang.org` is blocked while resolving `github.com/alparslanahmed/huidu-led v0.1.7`. The renderer, web UI, and config packages pass locally, and the Dockerfile retains the Portainer-safe `go build -mod=mod` behavior.
