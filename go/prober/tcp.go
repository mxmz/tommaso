package prober

import (
	"net"
	"time"
)

func TcpProbe(network string, timeout time.Duration) (time.Duration, error) {
	startTime := time.Now()
	conn, err := net.DialTimeout("tcp", network, timeout)
	endTime := time.Now()
	if err != nil {
		return 0, err
	} else {
		defer conn.Close()
		return endTime.Sub(startTime), nil
	}
}
