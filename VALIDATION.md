# Validation

Final simplified 64x32 display build.

- `go test ./internal/skywatch ./internal/webui ./internal/config` — PASS
- `go run ./cmd/previewgen` — PASS
- Generated frames are exactly 64x32 PNGs.
- Unknown/UFO artwork renders at 14x14; known operator artwork renders at 16x16.
- Callsign is the only large 5x7 text field; navigation and bottom telemetry use the compact 3x5 font.
- Text palette is limited to white, cyan, and amber.
- Dockerfile uses Go 1.25 Alpine and builds with `go build -mod=mod`.
- Dockerfile contains no executable `RUN go mod download` step.
- Compose retains configurable host port, PiAware URL, and external `adsb-net` network.

The complete Docker image build still depends on resolving `github.com/alparslanahmed/huidu-led` from the Docker host during `go build -mod=mod`.
