package main

import (
	"bufio"
	"flag"
	"net"
	"strings"
	"time"

	"post6.net/gohexdump/internal/drivers"
	"post6.net/gohexdump/internal/font"
	"post6.net/gohexdump/internal/screen"
)

func newRainScreen() screen.Screen {
	s := screen.NewHexScreen()
	s.SetFont(font.GetFont())
	filters := []screen.Filter{screen.NewRaindropFilter(s), screen.DefaultGamma()}
	return screen.NewFilterScreen(s, filters)
}

func newMessageScreen(message string) screen.Screen {
	conf := screen.Configuration{
		{0, 0, screen.HorizontalPanel},
		{0, 1, screen.HorizontalPanel},
		{0, 2, screen.HorizontalPanel},
		{0, 3, screen.HorizontalPanel},
	}
	s := screen.NewTextScreen(conf)
	s.SetFont(font.GetFont())
	s.SetStyle(screen.NewBrightness(1))
	s.WriteAt(strings.ToUpper(message), 0, 0)
	filters := []screen.Filter{screen.DefaultGamma(), screen.NewAfterGlowFilter(.85)}
	return screen.NewFilterScreen(s, filters)
}

func tcpListener(port string, screenChan chan<- screen.Screen, timeout time.Duration) {
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
				screenChan <- newRainScreen()
			}
		}(conn)
	}
}

func main() {
	port    := flag.String("port", "8080", "TCP port to listen on")
	webport := flag.String("webport", "8081", "HTTP port for web interface")
	timeout := flag.Duration("timeout", 30*time.Second, "time to show message before returning to rain")
	flag.Parse()

	refScreen := screen.NewHexScreen()
	refScreen.SetFont(font.GetFont())

	multi, screenChan := screen.NewMultiScreen()
	screenChan <- newRainScreen()

	go tcpListener(*port, screenChan, *timeout)
	go startWebServer(":"+*webport, screenChan, *timeout)

	q := make(chan bool)
	screen.DisplayRoutine(drivers.GetDriver(refScreen.SegmentCount()), multi, refScreen, q)
}
