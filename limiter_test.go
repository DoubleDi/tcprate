package tcprate

import (
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

const addr = "127.0.0.1:54321"

var data = []byte(strings.Repeat("d", 300))

// TestDynamicRate test transfers 300 bytes to a single conn and dynamicly changes the bandwith on server and conn
// first we set 15 bytes/second limit on conn
// after 10 seconds we set a server limit of 10 bytes/second
// finally, after 10 more seconds we set 5 bytes/second on conn
// total speed = 15 * 10 + 10 * 10 + 5 * 10 = 300 / 30 = 10 bytes/second
func TestDynamicRate(t *testing.T) {
	listener := runServer(t, func(limiter *Limiter) {
		go func() {
			<-time.After(time.Second * 10)
			t.Logf("changing server limit to %v", 10)
			limiter.WithBandwith(10)
		}()
	}, func(limiter *Limiter) {
		t.Logf("setting conn limit to %v", 15)
		limiter.WithPerConnBandwith(15)
		go func() {
			<-time.After(time.Second * 20)
			t.Logf("changing conn limit to %v", 5)
			limiter.WithPerConnBandwith(5)
		}()
	})
	defer listener.Close()
	runClient(t, func(speed float64) {
		if 9.5 > speed || 10.5 < speed {
			t.Errorf("incorrect speed %v", speed)
		}
	})
}

func runServer(t *testing.T, adjustLimiter func(limiter *Limiter), adjustConn func(limiter *Limiter)) net.Listener {
	limiter := NewLimiter()
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	adjustLimiter(limiter)
	adjustConn(limiter)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Log("closing server")
				return
			}
			wrappedConn := limiter.WrapConn(conn)
			go handleConn(t, wrappedConn)
		}
	}()
	return l
}

func handleConn(t *testing.T, c *Conn) {
	_, err := c.Write(data)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()
}

func runClient(t *testing.T, speedValidate func(speed float64)) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	now := time.Now()

	total := 0
	result := make([]byte, 1000)
	for {
		n, err := c.Read(result)
		since := time.Since(now)
		total += n
		speed := float64(total) / since.Seconds()
		t.Logf("got %v bytes in %v, speed avg %v bytes per second", total, since, speed)
		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
			speedValidate(speed)
			return
		default:
			t.Error(err)
			return
		}
	}
}

// TestMultipleClients test transfers 300 bytes to 2 conns and with a server limit 20 bytes/second
// total speed = 10 bytes/second per client
func TestMultipleClients(t *testing.T) {
	// start server
	listener := runServer(t, func(limiter *Limiter) {
		t.Logf("changing server limit to %v", 20)
		limiter.WithBandwith(20)
	}, func(limiter *Limiter) {})
	defer listener.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// start client 1
	go runClient(t, func(speed float64) {
		if 9.5 > speed || 10.5 < speed {
			t.Errorf("incorrect speed %v", speed)
		}
		wg.Done()
	})

	// start client 2
	runClient(t, func(speed float64) {
		if 9.5 > speed || 10.5 < speed {
			t.Errorf("incorrect speed %v", speed)
		}
		wg.Done()
	})

	wg.Wait()
}
