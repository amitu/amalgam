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
	var msg = []byte(fmt.Sprintf("%s.%s.:%v|ms", App, name, time/1000))
	count, err := statsdConn.Write(msg)
	if err != nil {
		return err
	}
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
