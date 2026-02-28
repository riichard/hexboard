package drivers

import (
	"flag"
	"os"
	"log"
	"post6.net/gohexdump/internal/util/clip"
	"github.com/jacobsa/go-serial/serial"
)

var baudrate uint
var serialDevice string

type Driver struct {
	file *os.File
	buf  []byte
}

func init() {
	flag.UintVar(&baudrate, "baudrate", 1500000, "serial baudrate")
	flag.StringVar(&serialDevice, "device", "/dev/ttyACM0", "serial output device")
}

func findActiveSerialDevice() string {
	options := serial.OpenOptions{
		BaudRate: baudrate,
		DataBits: 8,
		StopBits: 1,
		MinimumReadSize: 4,
	}

	// Example list of serial devices. In practice, you'd use a library function to get this list.
	devices := []string{"/dev/ttyACM0", "/dev/ttyACM1", "/dev/ttyUSB0"}

	for _, device := range devices {
		options.PortName = device
		if port, err := serial.Open(options); err == nil {
			// TODO: Perform a simple read/write test here if necessary.
			port.Close()
			return device // This device is available.
		}
	}
	log.Fatal("No active serial devices found")
	return ""
}

func GetDriver(size int) *Driver {
	serialDevice := findActiveSerialDevice()
	
	var err error
	size *= 2
	d := new(Driver)
	d.buf = make([]byte, size+4)
	d.buf[size] = 0xff
	d.buf[size+1] = 0xff
	d.buf[size+2] = 0xff
	d.buf[size+3] = 0xf0

	d.file, err = os.OpenFile(serialDevice, os.O_RDWR, 0)
	// Add error handling here for SetBaudrate and SetBinary if needed
	SetBaudrate(d.file, baudrate)
	SetBinary(d.file)

	if err != nil {
		panic("could not open serial device")
	}

	d.file.Write([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xf0}) // Discard frame

	return d
}

func (d *Driver) Write(data []float64) (int, error) {
	l := len(data)
	if l > (len(d.buf)-4)/2 {
		l = (len(d.buf)-4)/2
	}

	for i:=0; i<l ;i++ {
		v := clip.FloatToUintRange(data[i]*0xff00, 0, 0xff00)
		d.buf[i*2  ] = byte(v & 0xff)
		d.buf[i*2+1] = byte(v >> 8)
	}
	return d.file.Write(d.buf)
}

func (d *Driver) Close() error {
	return d.file.Close()
}

