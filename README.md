# ADS-B Matrix Display

A small Go service for a 64×32 RGB LED aircraft display driven by a Huidu HD-WF2 controller.

It reads live aircraft from PiAware/dump1090, finds the nearest eligible aircraft, classifies the operator (commercial, military, or unknown), renders a pixel-perfect 64×32 PNG, exposes a REST API, and can push the frame to the HD-WF2 over TCP.

## Architecture

```text
PiAware / dump1090
      │ aircraft.json
      ▼
adsb-matrix-display
      ├── REST API :8095
      ├── operator / military classification
      ├── 20×20 pixel assets
      └── 64×32 PNG renderer
              │ Wi-Fi/LAN TCP :10001
              ▼
           HD-WF2
              │ HUB75
              ▼
        64×32 P4 panel
```

The WF2 keeps its factory firmware. This service renders the entire screen image and sends it to the controller.

## What's included

- Nearest-aircraft calculation from receiver coordinates
- Position age, distance, altitude, and ground filtering
- Commercial airline detection from ICAO callsign prefixes
- Military detection from:
  - readsb `dbFlags` military bit when present
  - owner/operator metadata (`ownOp`) when present
  - the U.S. military ICAO allocation range used by readsb
  - a small conservative callsign fallback for selected U.S. military prefixes
- Branch-specific military icons for:
  - U.S. Air Force (`USAF`)
  - U.S. Marine Corps (`USMC`)
  - U.S. Coast Guard (`USCG`)
  - U.S. Navy (`USN`)
  - U.S. Army (`USA`)
  - generic military (`MIL`)
- Generic top-down aircraft icon for unknown/general-aviation traffic (`UNK`)
- Airline/regional icons for UAL, DAL, AAL, SWA, JBU, ASA, NKS, FFT, AAY, ACA, RPA, EDV, ENY, JIA/PSA, SKW, and AWI
- Browser/API preview of the exact 64×32 frame
- Manual and automatic HD-WF2 pushes
- Brightness and screen on/off controls

## Project layout

```text
.
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── README.md
├── VALIDATION.md
├── cmd/
│   ├── server/
│   └── previewgen/
├── design/
│   └── operator_pixel_asset_preview.png
├── examples/
│   └── frames/
│       ├── united.png
│       ├── delta.png
│       ├── southwest.png
│       ├── usaf.png
│       ├── military.png
│       ├── unknown.png
│       └── idle.png
└── internal/
    ├── assets/
    │   └── operators/*.png
    ├── config/
    ├── controller/
    ├── service/
    ├── skywatch/
    └── webui/
```

## Portainer deployment

This stack expects your existing external Docker network:

```text
adsb-net
```

The service resolves your PiAware container as `piaware` using:

```text
http://piaware:8080/data/aircraft.json
```

Set these Portainer stack environment variables (the first three are optional overrides):

```text
HOST_PORT=8095
ADSB_NETWORK=adsb-net
PIAWARE_URL=http://piaware:8080/data/aircraft.json
DISPLAY_NAME=ADS-B Matrix
TZ=UTC
RECEIVER_LAT=<receiver latitude>
RECEIVER_LON=<receiver longitude>
WF2_HOST=
PUSH_ENABLED=false
API_TOKEN=<optional token>
```

Leave `WF2_HOST` empty and `PUSH_ENABLED=false` until the physical display is configured and tested. Set `DISPLAY_NAME` privately in Portainer if you want custom station branding without committing it to Git.

Deploy `docker-compose.yml` as a Portainer stack. The Dockerfile intentionally does **not** run a separate `go mod download` layer. Module resolution happens during `go build -mod=mod`, using the public Go module proxy with a direct-source fallback. The builder uses Go 1.25 for compatibility with the current Huidu dependency.

The API will be available at:

```text
http://SERVER_IP:8095
```

## First API checks

```bash
curl http://SERVER_IP:8095/healthz
curl http://SERVER_IP:8095/api/v1/status
curl http://SERVER_IP:8095/api/v1/aircraft/nearest
curl http://SERVER_IP:8095/api/v1/display/preview
```

### Live browser display simulator

The service now includes a full browser-based simulator for the physical 64×32 panel. Open:

```text
http://SERVER_IP:8095/display
```

Opening the service root (`http://SERVER_IP:8095/`) redirects there automatically.

The simulator:

- shows the exact 64×32 PNG that will be sent to the HD-WF2
- enlarges it with nearest-neighbor/pixel rendering so every LED pixel stays crisp
- optionally overlays a 64×32 LED grid
- refreshes automatically every 2 seconds
- shows nearest-flight/operator, distance/direction, altitude/speed, tracked-aircraft count, and military count
- supports pause/resume and fullscreen mode
- works with `WF2_HOST` blank and `PUSH_ENABLED=false`, so no display hardware is required

You can still open the raw pixel-perfect frame directly at:

```text
http://SERVER_IP:8095/api/v1/display/frame.png
```

## Operator API

All available operator assets:

```text
GET /api/v1/operators
```

Commercial-only compatibility endpoint:

```text
GET /api/v1/airlines
```

Currently visible military aircraft:

```text
GET /api/v1/aircraft/military
```

Individual icon:

```text
GET /assets/operators/USAF.png
GET /assets/operators/UAL.png
GET /assets/operators/UNK.png
```

The old `/assets/airlines/{code}.png` route remains as a backwards-compatible alias.

## Example aircraft response

Commercial:

```json
{
  "hex": "A12345",
  "callsign": "UAL1827",
  "operator_code": "UAL",
  "operator_name": "United",
  "operator_kind": "airline",
  "military": false,
  "operator_detection": "callsign_prefix",
  "logo_url": "/assets/operators/UAL.png",
  "distance_nm": 2.8,
  "direction": "SW"
}
```

Military:

```json
{
  "hex": "AE1234",
  "callsign": "RCH421",
  "operator_code": "USAF",
  "operator_name": "United States Air Force",
  "operator_kind": "military",
  "military": true,
  "military_branch": "Air Force",
  "operator_detection": "military_flag+callsign",
  "logo_url": "/assets/operators/USAF.png"
}
```

## Military classification behavior

The service intentionally checks military classification before airline-prefix detection.

Priority:

1. `ownOp` / owner metadata identifies a U.S. military branch.
2. `dbFlags & 1` indicates military and the callsign can identify a branch.
3. Known U.S. military ICAO range identifies an aircraft as military; the callsign can optionally identify a branch.
4. Conservative military callsign heuristics.
5. Commercial airline prefix.
6. Unknown (`UNK`).

If the service knows an aircraft is military but cannot identify the branch confidently, it uses `MIL.png` instead of guessing.

### Why some aircraft will show generic `MIL`

Your current PiAware `aircraft.json` may not contain aircraft-database metadata such as `dbFlags` or `ownOp`. The service still recognizes the U.S. military ICAO range, but a hex address alone generally does not tell us Air Force vs Navy vs Army. In that case generic `MIL` is intentional.

A readsb/tar1090 source configured with an aircraft database can expose richer fields (`dbFlags`, registration, owner/operator) and improve branch resolution.

## HD-WF2 setup

First configure the screen normally using Huidu LedArt/HDSign so the WF2 knows the panel dimensions and scan parameters.

For the purchased panel, the expected configuration is:

```text
Width: 64
Height: 32
Color: Full color
Scan: 1/16
Output: HUB75E / 75EX1
```

Put the WF2 in Wi-Fi station mode and reserve its IP address in your router.

Then set:

```text
WF2_HOST=192.168.x.x
PUSH_ENABLED=false
```

### Manual test

```bash
curl -X POST http://SERVER_IP:8095/api/v1/display/test \
  -H 'Content-Type: application/json' \
  -d '{"text":"ADS-B MATRIX"}'
```

If `API_TOKEN` is configured:

```bash
-H 'Authorization: Bearer YOUR_TOKEN'
```

Push the current rendered aircraft frame:

```bash
curl -X POST http://SERVER_IP:8095/api/v1/display/push \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

After manual testing succeeds:

```text
PUSH_ENABLED=true
```

The service will then push only when the displayed information changes, with periodic forced refreshes.

## REST endpoints

```text
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

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `PIAWARE_URL` | `http://piaware:8080/data/aircraft.json` | ADS-B JSON feed |
| `RECEIVER_LAT` | required | Receiver latitude |
| `RECEIVER_LON` | required | Receiver longitude |
| `POLL_INTERVAL_SECONDS` | `2` | Feed polling interval |
| `MAX_POSITION_AGE_SECONDS` | `15` | Reject stale positions |
| `MAX_DISTANCE_NM` | `150` | Maximum display/search range |
| `MIN_ALTITUDE_FT` | `0` | Minimum altitude |
| `INCLUDE_GROUND` | `false` | Include aircraft on ground |
| `WF2_HOST` | empty | Controller LAN address |
| `WF2_PORT` | `10001` | Huidu SDK TCP port |
| `PUSH_ENABLED` | `false` | Automatically update display |
| `PUSH_INTERVAL_SECONDS` | `10` | Evaluate display push cadence |
| `FORCE_REFRESH_SECONDS` | `300` | Re-send unchanged frame periodically |
| `WF2_BRIGHTNESS` | `35` | Preferred brightness |
| `API_TOKEN` | empty | Optional POST authentication |

## Pixel assets

Every runtime icon in `internal/assets/operators/` remains a true 20×20 RGBA PNG source asset. The 64×32 renderer now downsizes the selected icon to a crisp 15×15 nearest-neighbor mark so aircraft identity no longer consumes nearly one-third of the screen width.

The aircraft frame is intentionally hierarchical:

- callsign is the primary 5×7 field
- operator/type sits on a compact secondary row
- distance and bearing get the larger middle row
- altitude and speed share a small explicit-unit telemetry strip (`6.1K FT`, `155 KT`)
- bearing uses a directional pixel arrow instead of a decorative dash line

`design/operator_pixel_asset_preview.png` is a nearest-neighbor enlarged preview of the embedded source assets. `design/redesign_preview.png` shows enlarged examples of the updated 64×32 layouts.

## Controller library

The controller package uses `github.com/alparslanahmed/huidu-led` and communicates with the HD-WF2 using Huidu SDK 2.0 over TCP port 10001. The rendered frame is uploaded as an in-memory PNG and referenced by a fullscreen image program.

## Public-repository privacy

This repository intentionally contains no receiver coordinates, controller IP, API token, feeder ID, or station-specific branding. Configure private deployment values through Portainer or an untracked `.env` file. Set `DISPLAY_NAME` privately if you want a custom station name on the idle screen and simulator UI.
