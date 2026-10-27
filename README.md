# ADS-B Matrix Display

A small self-hosted Go service that turns a live ADS-B feed into a 64×32 RGB LED aircraft display. It can run as a browser-only simulator before hardware arrives, then push the same pixel-perfect frame to a Huidu HD-WF2 controller driving a HUB75 panel.

The repository is intentionally deployment-neutral: receiver coordinates, controller addresses, API tokens, timezone, and display branding are supplied at runtime through environment variables and should not be committed to source control.

## What it does

- Reads `aircraft.json` from PiAware/dump1090/readsb-compatible sources.
- Filters stale or out-of-range positions.
- Selects the nearest eligible aircraft.
- Classifies commercial, military, and unknown operators.
- Detects U.S. military aircraft using database flags, owner/operator metadata, known ICAO ranges, and conservative callsign fallbacks.
- Uses branch-specific military assets for Air Force, Marines, Coast Guard, Navy, and Army.
- Renders a true 64×32 PNG frame with a 20×20 operator icon plus flight information.
- Serves a live browser simulator at `/display`.
- Pushes frames to an HD-WF2 over TCP when enabled.
- Exposes a small REST API for status, aircraft data, assets, preview frames, brightness, and screen control.

## Architecture

```text
ADS-B receiver / PiAware / readsb
            │ aircraft.json
            ▼
     ADS-B Matrix Display
       ├── REST API
       ├── operator classification
       ├── 20×20 pixel assets
       ├── 64×32 renderer
       └── browser simulator
                │
                │ optional TCP :10001
                ▼
             HD-WF2
                │ HUB75
                ▼
          64×32 RGB panel
```

## Browser simulator

No controller or LED panel is required to use the preview.

After deployment, open:

```text
http://SERVER_IP:8095/display
```

The page displays the exact 64×32 PNG that the hardware controller would receive, enlarged with nearest-neighbor rendering. It also shows flight/operator metadata, distance, altitude, speed, aircraft count, and visible military count.

The raw frame is available at:

```text
http://SERVER_IP:8095/api/v1/display/frame.png
```

## Public-safe configuration

Copy `.env.example` only as a reference. Keep real deployment values in Portainer, Docker secrets, or an untracked local `.env` file.

Required:

```text
RECEIVER_LAT=<receiver latitude>
RECEIVER_LON=<receiver longitude>
```

Useful optional settings:

```text
DISPLAY_NAME=ADS-B Matrix
TZ=UTC
PIAWARE_URL=http://piaware:8080/data/aircraft.json
WF2_HOST=
PUSH_ENABLED=false
API_TOKEN=
```

`DISPLAY_NAME` is runtime-only branding. For example, a private deployment can use any station name without adding it to the public repository.

## Portainer deployment

This project is designed to be deployed as a Git-backed Portainer Stack because the Compose file builds the image from the included Dockerfile and source tree.

By default it expects an existing external Docker network called:

```text
adsb-net
```

You can override that network name with `ADSB_NETWORK`.

Recommended Portainer environment variables:

```text
DISPLAY_NAME=ADS-B Matrix
TZ=UTC
RECEIVER_LAT=<receiver latitude>
RECEIVER_LON=<receiver longitude>
WF2_HOST=
PUSH_ENABLED=false
API_TOKEN=<optional long random token>
```

Leave `WF2_HOST` empty and `PUSH_ENABLED=false` until the physical controller has been configured and tested.

## First checks

```bash
curl http://SERVER_IP:8095/healthz
curl http://SERVER_IP:8095/api/v1/status
curl http://SERVER_IP:8095/api/v1/aircraft/nearest
curl http://SERVER_IP:8095/api/v1/display/preview
```

## HD-WF2 setup

Configure the panel in Huidu LedArt/HDSign first. A typical 64×32 P4 panel configuration is:

```text
Width: 64
Height: 32
Color: Full color
Scan: 1/16
Output: HUB75E / 75EX1
```

Put the WF2 in Wi-Fi station mode and give it a stable LAN address. Then set:

```text
WF2_HOST=<controller LAN address>
PUSH_ENABLED=false
```

Manual controller test:

```bash
curl -X POST http://SERVER_IP:8095/api/v1/display/test \
  -H 'Content-Type: application/json' \
  -d '{"text":"ADS-B MATRIX"}'
```

Push the current rendered frame:

```bash
curl -X POST http://SERVER_IP:8095/api/v1/display/push
```

If `API_TOKEN` is set, add:

```text
Authorization: Bearer YOUR_TOKEN
```

After manual testing succeeds:

```text
PUSH_ENABLED=true
```

## Operator and military classification

Classification priority is deliberately conservative:

1. Owner/operator metadata identifies a U.S. military branch.
2. `dbFlags & 1` identifies a military aircraft; callsign may refine the branch.
3. Known U.S. military ICAO allocation identifies military status.
4. A small set of characteristic callsign prefixes is used as a fallback.
5. Commercial ICAO callsign prefix is checked.
6. Anything unresolved uses the `UNK` asset.

If military status is known but the branch cannot be determined confidently, the generic `MIL` asset is used instead of guessing.

## Pixel assets

Runtime icons live in:

```text
internal/assets/operators/
```

Every runtime asset is a true 20×20 RGBA PNG and is placed directly into the left side of the 64×32 frame without resampling.

Included categories:

- Commercial and regional airlines
- U.S. Air Force (`USAF`)
- U.S. Marine Corps (`USMC`)
- U.S. Coast Guard (`USCG`)
- U.S. Navy (`USN`)
- U.S. Army (`USA`)
- Generic military (`MIL`)
- Unknown / mystery aircraft (`UNK`)

`design/operator_pixel_asset_preview.png` is an enlarged nearest-neighbor preview of the actual embedded assets.

## REST endpoints

```text
GET  /
GET  /display
GET  /healthz
GET  /api/v1/status
GET  /api/v1/aircraft
GET  /api/v1/aircraft/nearest
GET  /api/v1/aircraft/military
GET  /api/v1/operators
GET  /api/v1/airlines
GET  /api/v1/display/preview
GET  /api/v1/display/frame.png
GET  /assets/operators/{code}.png

POST /api/v1/display/test
POST /api/v1/display/push
POST /api/v1/controller/brightness?value=35
POST /api/v1/controller/screen?state=on
POST /api/v1/controller/screen?state=off
```

The old `/assets/airlines/{code}.png` route remains as a compatibility alias.

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `DISPLAY_NAME` | `ADS-B Matrix` | Runtime display/station name |
| `TZ` | `UTC` | Container timezone |
| `PIAWARE_URL` | `http://piaware:8080/data/aircraft.json` | ADS-B JSON source |
| `RECEIVER_LAT` | required | Receiver latitude |
| `RECEIVER_LON` | required | Receiver longitude |
| `POLL_INTERVAL_SECONDS` | `2` | Feed polling interval |
| `MAX_POSITION_AGE_SECONDS` | `15` | Reject stale positions |
| `MAX_DISTANCE_NM` | `150` | Maximum range |
| `MIN_ALTITUDE_FT` | `0` | Minimum altitude |
| `INCLUDE_GROUND` | `false` | Include aircraft on ground |
| `WF2_HOST` | empty | HD-WF2 LAN address |
| `WF2_PORT` | `10001` | Huidu SDK TCP port |
| `PUSH_ENABLED` | `false` | Automatically push frames |
| `PUSH_INTERVAL_SECONDS` | `10` | Auto-push evaluation interval |
| `FORCE_REFRESH_SECONDS` | `300` | Periodic unchanged-frame resend |
| `WF2_BRIGHTNESS` | `35` | Controller brightness |
| `API_TOKEN` | empty | Optional authentication for POST endpoints |
| `HOST_PORT` | `8095` | Host HTTP port in Compose |
| `ADSB_NETWORK` | `adsb-net` | External Docker network name |

## Security and privacy

This service is intended primarily for a trusted LAN.

- Do not commit real receiver coordinates, controller LAN addresses, tokens, or private station branding.
- `.env` files are ignored by Git.
- Avoid exposing the HTTP port directly to the public Internet. GET endpoints expose current receiver observations and status information.
- Use `API_TOKEN` to protect mutating POST endpoints.
- Keep your exact receiver location in deployment configuration rather than source code.

See `PRIVACY.md` for the public-repository scrub checklist used for this distribution.

## Development

Run the internal tests:

```bash
go test ./internal/skywatch ./internal/webui
```

Build:

```bash
go build ./cmd/server
```

The project uses `github.com/alparslanahmed/huidu-led` for HD-WF2 communication.

The module path in `go.mod` is deliberately neutral (`example.com/adsb-matrix-display`) so this public source package does not encode an individual's GitHub username. It can remain as-is for an application build, or be changed to the repository's final canonical path if desired.
