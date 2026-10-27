# Changelog

## Public-release privacy pass

- Renamed the public project to **ADS-B Matrix Display**.
- Removed station-specific branding and GitHub-user-specific module/import paths.
- Added runtime `DISPLAY_NAME` branding.
- Made timezone, PiAware URL, host port, and Docker network configurable.
- Removed the branded concept-art file from the distributable project.
- Added `PRIVACY.md` with public-repository guidance.
- Regenerated generic example frames.

## Browser display simulator

- Added `/display`, a live preview page for the exact 64×32 rendered frame.
- Added root redirect (`/` → `/display`).
- Added pixel-grid overlay, fullscreen mode, pause/resume, raw-frame link, and live flight/status metadata.
- Simulator works before the HD-WF2 hardware is configured.

## Airline + military operator assets

- Replaced airline-only assets with generic operator assets.
- Added true 20×20 pixel assets for commercial/regional carriers.
- Added U.S. Air Force, Marine Corps, Coast Guard, Navy, Army, and generic military assets.
- Added UFO-style unknown asset.
- Added military/operator classification to API models.
- Added `/api/v1/operators`, `/api/v1/aircraft/military`, and `/assets/operators/{code}.png`.
