package tcprate

import (
	"context"
	"net"

	"golang.org/x/time/rate"
)

type Conn struct {
	net.Conn
	base    *Limiter
	limiter *rate.Limiter
}

func (l *Limiter) WrapConn(c net.Conn) *Conn {
	return &Conn{
		Conn:    c,
		base:    l,
		limiter: rate.NewLimiter(rate.Inf, 1),
	}
}

func (c *Conn) WithBandwith(n int) *Conn {
	c.limiter.SetLimit(rate.Limit(n))
	return c
}

func (c *Conn) Write(b []byte) (n int, err error) {
	i := 0
	for i < len(b) {
		if err := c.limiter.WaitN(context.Background(), 1); err != nil {
			return n, err
		}
		if err := c.base.limiter.WaitN(context.Background(), 1); err != nil {
			return n, err
		}
		nn, err := c.Conn.Write(b[i : i+1])
		n += nn
		if err != nil {
			return n, err
		}
		i++
	}
	return n, nil
}
