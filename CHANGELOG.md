# Changelog

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
