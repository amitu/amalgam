package amalgam

import (
	"fmt"
	"net"
	"time"

	"github.com/juju/errors"
)

var statsdConn net.Conn

func statsdInit() {
	var err error
	fmt.Println("########")
	fmt.Println(StatsD)
	statsdConn, err = net.Dial("udp", StatsD)
	if err != nil {
		panic(err.Error())
	}
}

func Gauge(name string, value int) error {
	var msg = []byte(fmt.Sprintf("%s.%s.:%d|g", App, name, value))
	count, err := statsdConn.Write(msg)
	if err != nil {
		return err
	}
	if count != len(msg) {
		return errors.New("")
	}
	return nil
}

func Timer(name string, time time.Duration) error {
	fmt.Println("INSIDE TIMER")
	var msg = []byte("statsd.r2d2_event:100|ms")
	count, err := statsdConn.Write(msg)
	if err != nil {
		panic("Asdas")
		return err
	}
	fmt.Println("@@@@@@@@@")
	fmt.Println(count)
	if count != len(msg) {
		return errors.New("")
	}
	return nil
}

func Counter(name string, value int, sampling float64) error {
	var msg = []byte(fmt.Sprintf("%s.%s.:%d|c|@%f", App, name, value, sampling))
	count, err := statsdConn.Write(msg)
	if err != nil {
		return err
	}
	if count != len(msg) {
		return errors.New("")
	}
	return nil
}
