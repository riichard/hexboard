Hexboard
========

A wall-mounted 16-segment LED display driven by a Raspberry Pi. Shows a raindrop animation by default and switches to text when a TCP message arrives.

![hexboard](pics/hexboard.jpg)

## Hardware

- 4 rows × 32 columns of 16-segment LED digits
- Raspberry Pi (ARM) connected via serial (`/dev/ttyACM0`)
- Device hostname: `txt`

## Editor cursor integration

The display mirrors your editor cursor in real time using the rectripple effect — ripples radiate outward from wherever your cursor is.

**Neovim** — add to `init.lua`:

```lua
-- load from repo
vim.opt.runtimepath:append('/path/to/hexboard/editor')
require('hexboard').setup({ host = 'txt', port = 8082 })

-- or if you cloned the repo locally, just source the file:
-- require('/path/to/hexboard/editor/hexboard').setup()
```

The cursor position is mapped onto the 4×32 display grid:
- `col = editor_col % 32`
- `row = (editor_line - 1) % 4`

You can also set the cursor programmatically:

```bash
echo "10 2" | nc txt 8082          # TCP: col=10 row=2
curl -d "x=10&y=2" http://txt/cursor  # HTTP
```

## Quick start

First-time setup — installs and enables the systemd service so it starts on boot:

```bash
cd gohexdump
make install
```

After code changes:

```bash
make deploy   # sync + build + restart service
```

The display shows the raindrop animation. Open `http://txt` to send messages.

## Sending messages

### Web interface (recommended)

Open `http://txt` in a browser. Type up to 4 lines and tap **SEND**. Recent messages are listed below and can be re-sent with one tap.

Works on mobile.

### TCP (scripting)

Send any newline-terminated string over TCP to port 8080:

```bash
echo "hello world" | nc txt 8080
```

**From a script:**
```bash
echo "deploy complete" | nc txt 8080
```

---

The display shows the message (uppercased) for 30 seconds then returns to rain. Multiple messages sent in quick succession each start their own timer — the last one to expire wins.

**Custom timeout** (set at startup):
```bash
ssh txt 'nohup ~/hexboard -timeout 1m > /tmp/hexboard.log 2>&1 &'
```

## Makefile targets

Run from `gohexdump/`:

| Target | Description |
|---|---|
| `make install` | Sync + build + install systemd service (run once to set up) |
| `make deploy` | Sync + build + restart the running service |
| `make sync` | Rsync sources to device only |
| `make build` | Build binary on device only |
| `make run` | Interactive foreground session (SSH with TTY, bypasses systemd) |
| `make stop` | Stop the running process |
| `make status` | Show systemd service status |
| `make log` | Live journal output (`journalctl -fu hexboard`) |
| `make uninstall` | Remove the systemd service |
| `make test-message MSG="hi"` | Send a test message over TCP |

Override the target device with `DEVICE=`:
```bash
make deploy DEVICE=192.168.178.67
```

## `hexboard` flags

```
-device string      serial output device (default "/dev/ttyACM0")
-baudrate uint      serial baudrate (default 1500000)
-port string        TCP port for text messages (default "8080")
-webport string     HTTP port for web interface (default "80")
-cursorport string  TCP port for cursor position updates (default "8082")
-timeout duration   time to show message before returning to idle (default 30s)
-verbose            print FPS to stdout
```

## Other commands

All commands connect to the serial device and run on the `txt` server.

### `raindrops`

Standalone raindrop animation with no TCP server.

```bash
ssh txt 'cd ~/dev/hexboard/gohexdump && /usr/local/go/bin/go run ./cmd/raindrops'
```

### `playvid`

Play a pre-encoded video file on the display. Reads raw segment frames from stdin.

```bash
ssh txt 'cat ~/matrix.bin | ~/playvid'
ssh txt '~/playvid -fps 24 < ~/intro.bin'
```

Flags: `-fps int` (default 30), `-gamma float64` (default 2.5)

### `encvid`

Encode a raw video stream (grayscale pixels) into segment frames for `playvid`. Reads raw video from stdin, writes segment frames to stdout.

```bash
ffmpeg -i input.mp4 -vf scale=1280:720 -pix_fmt gray -f rawvideo - \
  | ./encvid -width 1280 -height 720 > output.bin
```

Flags: `-width int` (default 1280), `-height int` (default 720)

## Project structure

```
gohexdump/
  cmd/
    hexboard/     # main program: rain + TCP message mode
    raindrops/    # standalone rain animation
    playvid/      # video playback
    encvid/       # video encoder (run locally, output copied to device)
  internal/
    drivers/      # serial driver (CGo, Linux only)
    font/         # 16-segment font
    screen/       # display abstractions (TextScreen, filters, animation)
    tcpserver/    # legacy TCP keyboard input (unused by hexboard)
```

## Building

The `drivers` package uses Linux CGo (`asm/termbits.h`), so the binary must be built on the device itself. The Makefile handles this via SSH.

To build manually on the device:
```bash
ssh txt 'cd ~/dev/hexboard/gohexdump && /usr/local/go/bin/go build -o ~/hexboard ./cmd/hexboard'
```
