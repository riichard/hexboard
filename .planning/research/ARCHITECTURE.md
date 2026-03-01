# Architecture Research: Hexboard Web UI Enhancement

**Research type:** Architecture — subsequent milestone
**Date:** 2026-03-01

---

## Existing Architecture (relevant parts)

```
cmd/hexboard/
  main.go       — flag.Parse, serial init, screenChan, DisplayRoutine (blocks)
  web.go        — HTTP server port 80; GET / → HTML form; POST / → send(msg)

internal/screen/
  screen.go     — Screen interface: NextFrame(f, old *FrameBuffer, tick uint64) bool
  text.go       — TextScreen (WriteAt), HexScreen, FilterScreen, MultiScreen
  display.go    — DisplayRoutine (60fps render loop)

internal/font/
  font.go       — 16-segment font table

drivers/        — CGo serial driver
```

**Key constraint:** `DisplayRoutine` blocks the main goroutine. All I/O (HTTP, Hue, DB) must run in separate goroutines.

---

## New Components

### 1. `internal/store/` — SQLite history store

```go
package store

type Entry struct {
    ID        int64
    Type      string    // "message" | "drawing"
    Content   string    // JSON: {lines: [...]} or {segments: [...]}
    Duration  *int      // nil = use global default
    CreatedAt time.Time
}

type Store struct { db *sql.DB }

func Open(path string) (*Store, error)
func (s *Store) Save(entry Entry) (int64, error)
func (s *Store) List(limit int) ([]Entry, error)
func (s *Store) Delete(id int64) error
```

**File location:** `/var/lib/hexboard/history.db` (configurable via flag)

---

### 2. `internal/hue/` — Hue Bridge client

```go
package hue

type Config struct {
    BridgeIP string
    APIKey   string
    DeviceID string
}

type Client struct { cfg Config; http *http.Client }

func NewClient(cfg Config) *Client
func (c *Client) TurnOn() error   // PUT /api/<key>/lights/<id>/state {"on":true}
```

**Called from web.go** in a goroutine (fire-and-forget, errors logged not propagated).

---

### 3. Config file — `hexboard.conf`

```toml
# /etc/hexboard.conf or ~/hexboard.conf
[hue]
bridge_ip = ""
api_key   = ""
device_id = ""

[display]
default_duration_sec = 30
```

Read at startup in `main.go`. Flags override config file values.

---

### 4. `cmd/hexboard/web.go` — Extended

New HTTP routes:
- `GET /history` — JSON array of history entries
- `POST /draw` — accepts segment bitmask array, saves to store, pushes DrawingScreen to screenChan
- `GET /settings` — return current global settings (JSON)
- `POST /settings` — update global default duration

Existing routes extended:
- `POST /` — now saves to store, triggers Hue.TurnOn() in goroutine, respects per-message duration

---

### 5. Browser: Drawing Tool

New HTML page or section in existing `/` page:
- SVG grid of 4×32 digit cells, each with 16 clickable segment paths
- JS state: `uint16[]` of length 128
- On submit: POST to `/draw` with JSON body `{segments: [...], duration: N}`
- Re-uses existing HTML template system

---

## Data Flow: Send a Drawing

```
User clicks segment in browser SVG
  → JS toggles bit in segments[col][row]
  → User sets duration (or uses global default)
  → POST /draw {segments: [...128 uint16s...], duration: 30}
    → store.Save(Entry{Type:"drawing", Content:json, Duration:&30})
    → NewDrawingScreen(segments) pushed to screenChan
    → go hue.TurnOn()  (non-blocking)
    → HTTP 200 with updated history list
```

## Data Flow: Re-send from History

```
User clicks history item
  → Browser shows duration prompt (default: stored duration or global default)
  → POST / or POST /draw with content + duration
    → Same path as new send
```

---

## Build Order

1. **store package** — SQLite schema, CRUD (no UI dependencies)
2. **hue package** — HTTP client (no UI dependencies)
3. **Config file loading** — extend main.go
4. **Drawing screen** — new Screen impl that renders segment bitmasks
5. **Extended web routes** — history API, draw endpoint
6. **Browser drawing tool** — SVG segment picker
7. **Configurable timeout** — plumb duration through send → screen switch → timer
8. **History UI** — replace existing recent-messages list with DB-backed list
