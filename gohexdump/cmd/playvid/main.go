
package main

import (
	"post6.net/gohexdump/internal/drivers"
	"os"
	"io"
	"time"
	"flag"
	"math"
)

var fps int
var gamma float64

func init() {

	flag.IntVar(&fps, "fps", 30, "fps")
	flag.Float64Var(&gamma, "gamma", 2.5, "gamma")
}

func main() {

	flag.Parse()

	var gmap [256]float64
	var buf     [960*16]byte
	var buf_f64 [960*16]float64

	for i := range gmap {
		v := math.Pow(float64(i)/255, gamma)
		if v > 1 {
			v = 1
		} else if v < 0 {
			v = 0
		}
		gmap[i] = v
	}

	out := drivers.GetDriver(len(buf))

	tick := time.NewTicker(time.Second / time.Duration(fps))
	for {
		if _, err := io.ReadFull(os.Stdin, buf[:]); err != nil {
			break
		}

		for i := range(buf) {
			buf_f64[i] = gmap[buf[i]]
		}

		out.Write(buf_f64[:])

		<-tick.C
	}
	for i := range(buf) {
		buf_f64[i] = 0
	}
	out.Write(buf_f64[:])
}

