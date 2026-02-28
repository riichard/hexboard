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
make deploy    # sync + build + restart
make run       # foreground interactive session
```

Code is synced via `rsync` (no GitHub SSH key on device). After local changes, always `make deploy` or at minimum `make sync && make build`.

## Main program: `cmd/hexboard`

- Shows raindrop animation by default
- Switches to uppercase text on TCP message (port 8080, newline-terminated)
- Returns to rain after `-timeout` (default 30s)
- `screen.MultiScreen` handles switching: last value pushed to `screenChan` wins

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

## Sending a test message

```bash
echo "hello" | nc txt 8080
# or
make test-message MSG="hello" -C gohexdump
```
