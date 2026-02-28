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

// identity transform — ripples radiate outward in screen-space (rectripple style)
func identityTransform(v screen.Vector2) screen.Vector2 { return v }

// newDefaultScreen returns the idle rectripple screen and its cursor.
// The same pair is reused for the lifetime of the process so cursor
// position is preserved across message display cycles.
func newDefaultScreen() (screen.Screen, screen.Cursor) {
	s := screen.NewTextScreen(hexConf)
	s.SetFont(font.GetFont())
	s.SetStyle(screen.NewBrightness(1))
	cursor := screen.NewRippleCursor(1, .5, nil, identityTransform, s)
	filters := []screen.Filter{cursor, screen.DefaultGamma(), screen.NewAfterGlowFilter(.85)}
	return screen.NewFilterScreen(s, filters), cursor
}

func newMessageScreen(message string) screen.Screen {
	s := screen.NewTextScreen(hexConf)
	s.SetFont(font.GetFont())
	s.SetStyle(screen.NewBrightness(1))
	for row, line := range strings.SplitN(strings.ToUpper(message), "\n", 4) {
		runes := []rune(line)
		if len(runes) > 32 {
			runes = runes[:32]
		}
		s.WriteAt(string(runes), 0, row)
	}
	filters := []screen.Filter{screen.DefaultGamma(), screen.NewAfterGlowFilter(.85)}
	return screen.NewFilterScreen(s, filters)
}

func tcpListener(port string, screenChan chan<- screen.Screen, idleScreen screen.Screen, timeout time.Duration) {
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
				msg := scanner.Text()
				screenChan <- newMessageScreen(msg)
				time.Sleep(timeout)
				screenChan <- idleScreen
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

	idleScreen, cursor := newDefaultScreen()
	cursor.SetCursor(0, 0)

	multi, screenChan := screen.NewMultiScreen()
	screenChan <- idleScreen

	go tcpListener(*port, screenChan, idleScreen, *timeout)
	go cursorListener(*cursorport, cursor)
	go startWebServer(":"+*webport, screenChan, idleScreen, cursor, *timeout)

	q := make(chan bool)
	screen.DisplayRoutine(drivers.GetDriver(refScreen.SegmentCount()), multi, refScreen, q)
}
