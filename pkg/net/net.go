package net

import (
	"net"
	"time"
)

func IsReachableHost(host, port string, timeout time.Duration) bool {
	return DialPort(host, port, timeout)
}

func DialPort(host, port string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()

	return true
}
