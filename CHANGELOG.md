
## Deployment fix

- Restored the Portainer-safe Docker build command using `go build -mod=mod`.
- Restored `GOPROXY=https://proxy.golang.org|direct` for resilient module resolution.
- Restored configurable `HOST_PORT`, `PIAWARE_URL`, and external `ADSB_NETWORK` values in Compose.
# Changelog

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
