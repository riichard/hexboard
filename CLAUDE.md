# Hexboard — Claude context

## What this project is

A wall-mounted 16-segment LED display driven by a Raspberry Pi Zero (ARMv6). The Go code in `gohexdump/` runs on the Pi and drives the display over serial.

## Key facts

- **Device hostname:** `txt` (Raspberry Pi, 192.168.178.67, Linux ARMv6)
- **Go on device:** `/usr/local/go/bin/go` (Go 1.19, linux/arm)
- **Binary lives at:** `~/hexboard` on the device
- **Serial port:** `/dev/ttyACM0` at 1500000 baud
- **Display:** 4 rows × 32 columns of 16-segment digits (128 digits, 2048 segments total)
- **Module path:** `post6.net/gohexdump`

## Build and deploy

The `drivers` package uses Linux CGo (`asm/termbits.h`) — **build on the device, not locally**.

```bash
cd gohexdump
make install   # first-time: sync + build + install systemd service
make deploy    # after changes: sync + build + restart service
make run       # foreground interactive session (bypasses systemd)
```

Code is synced via `rsync` (no GitHub SSH key on device). After local changes, always `make deploy` or at minimum `make sync && make build`.

## Main program: `cmd/hexboard`

- **Default (idle) mode:** rectripple cursor effect — blank screen with rippling cursor, using identity transform (ripples propagate in screen-space)
- **Message mode:** switches to uppercase text on TCP message (port 8080) or web POST
- Returns to idle (`idleScreen`) after `-timeout` (default 30s) — the **same** screen object is reused so cursor position persists
- **Cursor port (8082):** accepts persistent TCP connections; each line `col row\n` calls `cursor.SetCursor(col, row)`
- **HTTP `/cursor`:** `POST x=<col>&y=<row>` (204 No Content)
- `screen.MultiScreen` handles switching: last value pushed to `screenChan` wins

### Key design: long-lived idle screen

`idleScreen` and `cursor` are created once in `main()` and shared across all goroutines. When returning from message mode, goroutines push `idleScreen` back onto `screenChan` (not a freshly constructed screen). This preserves cursor position across display switches.

## Architecture: `internal/screen`

- `Screen` interface: `NextFrame(f, old *FrameBuffer, tick uint64) bool`
- `TextScreen`: grid of 16-segment digits, writable with `WriteAt(string, col, row)`
- `HexScreen`: 4×32 `TextScreen` using `HorizontalPanel` layout
- `FilterScreen`: wraps a `Screen` and applies a chain of `Filter`s per frame
- `MultiScreen`: round-robins a channel of `Screen` values; used for hot-swapping screens
- `DisplayRoutine`: 60fps render loop; runs in the **main goroutine** (blocks)
- `FrameBuffer`: flat `[]float64` of `DigitCount * 16` segment brightnesses

### Screen construction pattern

```go
s := screen.NewHexScreen()
s.SetFont(font.GetFont())
filters := []screen.Filter{screen.NewRaindropFilter(s), screen.DefaultGamma()}
screen.NewFilterScreen(s, filters)
```

### Adding a new display mode

Create a function returning `screen.Screen` and push it onto `screenChan`:
```go
screenChan <- newMyScreen()
```

## Known quirks

- `drivers.init()` registers `-device` and `-baudrate` flags but does **not** call `flag.Parse()` — callers must call `flag.Parse()` in `main()`
- `FrameBuffer.frame` has exactly `DigitCount * 16` entries (16 segments per digit). Raindrop rendering iterates segment bits 0–15 only
- Many raindrop columns (>31) are outside the HexScreen grid and silently no-op via `DigitIndex` returning -1
- `NewRaindropFilter` takes a `TextScreen` — pass the inner `HexScreen`, not the `FilterScreen` wrapper

## Web interface

`cmd/hexboard/web.go` — HTTP server on port 80 (flag: `-webport`), runs as root via systemd.

- `GET /` — renders the HTML form with a recent-messages list
- `POST /` — calls `send(msg)` which pushes to `screenChan` and starts a rain-return timer in a goroutine

Multi-line messages: newlines in the form textarea are split by `\n` in `newMessageScreen`; each line is written to its own row (0–3), truncated to 32 chars. Recent messages kept in memory (last 10). Each recent item is a hidden-input form so clicking re-sends with no JS.

## systemd

Service file: `gohexdump/hexboard.service` — deployed to `/etc/systemd/system/hexboard.service`.
Runs as `root` (needed for port 80 and `/dev/ttyACM0`). Enabled at boot.

```bash
make install    # deploy service file, enable, start
make status     # systemctl status hexboard
make log        # journalctl -fu hexboard
make uninstall  # stop, disable, remove service file
```

## Sending a test message

```bash
echo "hello" | nc txt 8080                       # TCP text message
printf "line1\nline2" | nc txt 8080              # TCP multi-line
curl -d "message=hello" http://txt/              # HTTP text message
open http://txt                                  # web UI

echo "10 2" | nc txt 8082                        # TCP cursor: col=10 row=2
curl -d "x=10&y=2" http://txt/cursor             # HTTP cursor
```

## Editor plugin

`editor/hexboard.lua` — Neovim Lua plugin. Connects persistently to port 8082 on `CursorMoved`/`CursorMovedI`, sends `col row\n`, debounced to 40 ms. Reconnects on `FocusGained`.
