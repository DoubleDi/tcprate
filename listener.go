package tcprate

import (
	"net"
)

type Listener struct {
	net.Listener
	limiter *Limiter
}

func NewListener(l net.Listener) *Listener {
	return &Listener{
		Listener: l,
		limiter:  NewLimiter(),
	}
}
func (l *Listener) SetLimits(limitGlobal, limitPerConn int) {
	l.limiter.WithBandwith(limitGlobal)
	l.limiter.WithPerConnBandwith(limitPerConn)
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	return l.limiter.WrapConn(conn), err
}
