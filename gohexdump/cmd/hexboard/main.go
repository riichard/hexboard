package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"post6.net/gohexdump/internal/drivers"
	"post6.net/gohexdump/internal/font"
	"post6.net/gohexdump/internal/screen"
)

var hexConf = screen.Configuration{
	{0, 0, screen.HorizontalPanel},
	{0, 1, screen.HorizontalPanel},
	{0, 2, screen.HorizontalPanel},
	{0, 3, screen.HorizontalPanel},
}

// identity transform — ripples radiate in screen-space (same as rectripple)
func identityTransform(v screen.Vector2) screen.Vector2 { return v }

// display holds the two long-lived screen objects and the shared cursor.
//
//   rain   — matrix raindrop animation shown when idle
//   ripple — rectripple screen used when a message is displayed;
//            text is written into its text layer on each message
//   cursor — shared RippleCursor; editor updates always go here
type display struct {
	rain   screen.Screen
	ripple screen.Screen
	text   screen.TextScreen
	cursor screen.Cursor
}

func newDisplay() *display {
	// Idle: matrix rain
	hex := screen.NewHexScreen()
	hex.SetFont(font.GetFont())
	rain := screen.NewFilterScreen(hex, []screen.Filter{
		screen.NewRaindropFilter(hex),
		screen.DefaultGamma(),
	})

	// Text display: rectripple with cursor
	s := screen.NewTextScreen(hexConf)
	s.SetFont(font.GetFont())
	s.SetStyle(screen.NewBrightness(1))
	cursor := screen.NewRippleCursor(1, .5, nil, identityTransform, s)
	ripple := screen.NewFilterScreen(s, []screen.Filter{
		cursor,
		screen.DefaultGamma(),
		screen.NewAfterGlowFilter(.85),
	})

	return &display{rain: rain, ripple: ripple, text: s, cursor: cursor}
}

// showMessage writes msg into the rectripple text layer, switches to it,
// then returns to rain after timeout. Safe to call from multiple goroutines.
func (d *display) showMessage(msg string, screenChan chan<- screen.Screen, timeout time.Duration) {
	d.text.Clear()
	for row, line := range strings.SplitN(strings.ToUpper(msg), "\n", 4) {
		runes := []rune(line)
		if len(runes) > 32 {
			runes = runes[:32]
		}
		d.text.WriteAt(string(runes), 0, row)
	}
	screenChan <- d.ripple
	go func() {
		time.Sleep(timeout)
		screenChan <- d.rain
	}()
}

func tcpListener(port string, screenChan chan<- screen.Screen, d *display, timeout time.Duration) {
	listen, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			if scanner.Scan() {
				d.showMessage(scanner.Text(), screenChan, timeout)
			}
		}(conn)
	}
}

// cursorListener accepts persistent TCP connections and reads "col row\n"
// lines to update the cursor position in real time (e.g. from an editor).
func cursorListener(port string, cursor screen.Cursor) {
	listen, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		go func(conn net.Conn) {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				var col, row int
				if _, err := fmt.Sscan(scanner.Text(), &col, &row); err == nil {
					cursor.SetCursor(col, row)
				}
			}
		}(conn)
	}
}

func main() {
	port       := flag.String("port", "8080", "TCP port for text messages")
	webport    := flag.String("webport", "80", "HTTP port for web interface")
	cursorport := flag.String("cursorport", "8082", "TCP port for cursor position updates (col row\\n)")
	timeout    := flag.Duration("timeout", 30*time.Second, "time to show message before returning to idle")
	flag.Parse()

	refScreen := screen.NewHexScreen()
	refScreen.SetFont(font.GetFont())

	d := newDisplay()
	d.cursor.SetCursor(0, 0)

	multi, screenChan := screen.NewMultiScreen()
	screenChan <- d.rain

	go tcpListener(*port, screenChan, d, *timeout)
	go cursorListener(*cursorport, d.cursor)
	go startWebServer(":"+*webport, screenChan, d, *timeout)

	q := make(chan bool)
	screen.DisplayRoutine(drivers.GetDriver(refScreen.SegmentCount()), multi, refScreen, q)
}
