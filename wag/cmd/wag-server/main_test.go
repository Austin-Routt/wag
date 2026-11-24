package main

import (
	"net"
	"testing"
)

// A server that listens on every interface is opened as localhost, which it
// takes for its owner; any other address is opened as it is.
func TestPageAddr(t *testing.T) {
	for _, c := range []struct {
		addr net.Addr
		want string
	}{
		{&net.TCPAddr{IP: net.IPv4zero, Port: 8080}, "localhost:8080"},
		{&net.TCPAddr{IP: net.IPv6unspecified, Port: 8080}, "localhost:8080"},
		{&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9090}, "127.0.0.1:9090"},
		{&net.TCPAddr{IP: net.IPv4(192, 0, 2, 7), Port: 8080}, "192.0.2.7:8080"},
	} {
		if got := pageAddr(c.addr); got != c.want {
			t.Errorf("%v: %s, want %s", c.addr, got, c.want)
		}
	}
}
