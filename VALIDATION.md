# Validation

## Completed in this build

- `go test ./internal/skywatch` passes.
- Airline prefix resolution tested.
- Military resolution from `dbFlags` tested.
- U.S. military ICAO-range fallback tested.
- Branch classification from conservative callsign heuristics tested.
- 64×32 commercial render tested.
- 64×32 military render tested.
- All runtime operator source assets are exactly 20×20 PNGs.
- The frame renderer scales operator artwork to a crisp 15×15 nearest-neighbor footprint.
- Example commercial, military, unknown, and idle frames were regenerated from the current renderer.
- Updated telemetry formatting uses explicit units (`6.1K FT`, `155 KT`) and the distance/bearing row uses an 8-way pixel arrow.
- Huidu controller API usage was checked against the current upstream project documentation/source for:
  - `UploadFileData`
  - `FileTypeImage`
  - `AddProgramWithConfig`
  - `ImageFitStretch`
  - `WithAutoReconnect`
  - screen/brightness controls

## Environment limitation

This execution environment cannot reach `proxy.golang.org`, so `go test ./...` cannot download `github.com/alparslanahmed/huidu-led v0.1.7` and therefore cannot complete the controller-package compile here.

The Dockerfile intentionally avoids a standalone `go mod download` step. The production image resolves the pinned dependency as part of `go build -mod=mod`, so Portainer does not fail in a separate pre-build dependency layer.

## Military-data note

readsb documents `dbFlags & 1` as the military bit when an aircraft database is enabled. readsb also identifies known military ICAO ranges, including the U.S. allocation used by this project as a fallback.

Branch-specific classification is best-effort. When military status is known but branch identity is not, the service intentionally uses the generic `MIL` asset rather than inventing a branch.

## Browser simulator validation

Validated after adding the built-in `/display` page:

- `go test ./internal/webui ./internal/skywatch` passes.
- Full project compiles/tests against the local Huidu API compatibility stub.
- End-to-end smoke test with a mock PiAware feed passed:
  - `/` returns a 307 redirect to `/display`.
  - `/display` serves the embedded simulator HTML.
  - `/api/v1/status` reported two eligible aircraft and selected the nearest aircraft.
  - `/api/v1/aircraft/military` correctly returned the mock USAF aircraft.
  - `/api/v1/display/frame.png` returned a 64×32 RGB PNG.
  - `/assets/operators/UAL.png` returned a 20×20 RGBA PNG.
- Fixed a Go `http.ServeMux` runtime issue in the previous asset route pattern (`/{code}.png`); the new `{file}` route parses the `.png` suffix safely and starts correctly on modern Go.

The real `huidu-led` module still cannot be downloaded in this execution environment because outbound Go module access is blocked; production Docker builds continue to use the pinned real dependency, resolved during `go build -mod=mod`.
