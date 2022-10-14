package tcprate

import (
	"context"
	"net"
	"sync"

	"golang.org/x/time/rate"
)

type Conn struct {
	net.Conn
	mu      sync.Mutex
	base    *Limiter
	limiter *rate.Limiter
	limit   int
}

func (l *Limiter) WrapConn(c net.Conn) *Conn {
	perConn := l.perConnLimit()
	return &Conn{
		Conn:    c,
		base:    l,
		limiter: rate.NewLimiter(rate.Limit(perConn), perConn),
		limit:   perConn,
	}
}

func (c *Conn) withBandwith(n int) *Conn {
	c.limit = n
	c.limiter.SetLimit(rate.Limit(n))
	c.limiter.SetBurst(int(n))
	return c
}

func (c *Conn) Write(b []byte) (n int, err error) {
	return c.do(b, c.Conn.Write)
}

func (c *Conn) Read(b []byte) (n int, err error) {
	return c.do(b, c.Conn.Read)
}

func (c *Conn) do(b []byte, action func(b []byte) (n int, err error)) (n int, err error) {
	i := 0
	for i < len(b) {
		c.mu.Lock()
		if perConn := c.base.perConnLimit(); c.limit != perConn {
			c.withBandwith(perConn)
		}
		c.mu.Unlock()

		limit, err := c.base.BatchAndWait(len(b[i:]))
		if err != nil {
			return n, err
		}

		if err := c.limiter.WaitN(context.Background(), limit); err != nil {
			return n, err
		}

		nn, err := action(b[i : i+limit])
		n += nn
		if err != nil {
			return n, err
		}
		i += limit
	}
	return n, nil
}
