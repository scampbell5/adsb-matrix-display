# Validation

## Functional checks

- `go test ./internal/skywatch ./internal/webui` passes with network access disabled.
- Full `go test ./...` passes against a local compatibility implementation of the HD-WF2 library API.
- Airline-prefix resolution is tested.
- Military resolution from `dbFlags` is tested.
- U.S. military ICAO-range fallback is tested.
- Branch classification from conservative callsign heuristics is tested.
- Commercial and military 64×32 rendering is tested.
- Runtime operator assets are true 20×20 PNGs.
- Generic example frames were regenerated from the current renderer.

## Public-release scrub

The release tree was searched for station-specific branding, personal names, addresses, contact information, exact receiver coordinates, feeder IDs/UUIDs, private IP ranges, tokens, and the prior GitHub-user-specific module path.

Deployment-sensitive values are represented only by environment-variable placeholders.

PNG files were checked for embedded textual metadata before packaging.

## External dependency note

The production Docker build uses the pinned `github.com/alparslanahmed/huidu-led` dependency. In an execution environment without outbound Go module access, the full project was compiled/tested using a local API-compatible stub; the source tree itself retains the real dependency and no stub is included in the release ZIP.

## Runtime-branding smoke test

A local mock `aircraft.json` feed was served with no controller configured. The service was started with a non-default private runtime name and verified to:

- expose that name through `/api/v1/status`
- select the expected nearest aircraft
- serve `/display`
- return a valid PNG from `/api/v1/display/frame.png`
- keep the runtime-only private name out of the source tree
