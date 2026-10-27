# Deployment fix v3

- Removed the standalone `RUN go mod download` Docker layer that could fail before the existing `-mod=mod` workaround was reached.
- Dependency resolution now happens only during `go build -mod=mod`.
- Updated the builder image from Go 1.23 to Go 1.25 and set `GOTOOLCHAIN=local` to avoid an unexpected toolchain download during Portainer builds.
- Kept `GOPROXY=https://proxy.golang.org|direct` so a proxy error can fall back to the module source.


## Deployment fix

- Restored the Portainer-safe Docker build command using `go build -mod=mod`.
- Restored `GOPROXY=https://proxy.golang.org|direct` for resilient module resolution.
- Restored configurable `HOST_PORT`, `PIAWARE_URL`, and external `ADSB_NETWORK` values in Compose.
# Changelog

## 2026-10-27 — Final hybrid display pass

- Restored the original UFO-style unknown-aircraft asset.
- Increased on-screen operator artwork to 18×18 for more personality without returning to the original 20×20-heavy composition.
- Removed the generic `AIRCRAFT` filler label for unknown/general-aviation targets.
- Kept the callsign as the primary identity, while tightening the rest of the frame around a compact `distance + bearing` row and explicit `FT` / `KT` telemetry.
- Airline and military frames retain their operator logo and show operator/type context only when useful.
- Added updated `N1410A` and `N380F` example frames and refreshed the enlarged design preview.
- Preserved the v3 Portainer Docker fix: no standalone `go mod download`; dependency resolution remains in `go build -mod=mod`.

## 2026-10-27 — 64×32 display polish

- Reworked the aircraft frame around a stronger visual hierarchy for the physical P4 panel.
- Reduced operator artwork from a 20×20 on-screen footprint to a crisp 15×15 nearest-neighbor render.
- Made the callsign the primary white identity field and preserved up to eight characters at full 5×7 size.
- Added a compact operator/type identity row.
- Promoted distance and bearing to the middle row and replaced the decorative dash with a real 8-way direction arrow.
- Replaced compressed altitude text such as `6.1KF` with explicit units such as `6.1K FT`.
- Right-aligned speed as an explicit `KT` field.
- Replaced the UFO unknown mark with a generic top-down aircraft icon.
- Reworked the idle frame with centered branding and the generic aircraft mark.
- Updated example frames and design previews, including a comparison-friendly `N380F` unknown-aircraft sample.

## Browser display simulator

- Added `/display`, a live hardware-preview page for the exact 64×32 rendered frame.
- Added root redirect (`/` → `/display`).
- Added pixel-grid overlay, fullscreen mode, pause/resume, raw-frame link, and live flight/status metadata.
- Simulator works in API-only mode before the HD-WF2 hardware arrives.

## 2026-10-27 — Airline + Military operator assets

- Replaced airline-only asset model with generic operator assets.
- Added true 20×20 pixel assets for commercial/regional carriers.
- Added U.S. Air Force, Marine Corps, Coast Guard, Navy, Army, and generic military assets.
- Replaced the old unknown mark with a UFO-style unknown asset.
- Added operator kind (`airline`, `military`, `unknown`) to API models.
- Added military branch and detection method to aircraft/API responses.
- Added military detection using `dbFlags`, owner metadata, U.S. military ICAO range, and conservative callsign heuristics.
- Added `/api/v1/operators`.
- Added `/api/v1/aircraft/military`.
- Added `/assets/operators/{code}.png` while retaining the old airline asset route as an alias.
- Updated 64×32 frame rendering to use operator assets.
- Added USAF, generic military, and UFO unknown example frames.
- Added classification/render tests.

## 2026-10-27 - balanced 64x32 layout
- Reduced unknown/UFO artwork to a true 14x14 render size; known operator logos render at 16x16.
- Moved the callsign left to x=17 to visually pair it with the artwork.
- Rebuilt the navigation row as distance + compact NM + bearing arrow/direction.
- Rebuilt the bottom row with 5x7 altitude/speed values and compact 3x5 units.
- Preserved the Portainer-safe Docker build with no standalone go mod download step.
