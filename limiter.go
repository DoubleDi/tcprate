package tcprate

import (
	"golang.org/x/time/rate"
)

type Limiter struct {
	limiter *rate.Limiter
}

func NewLimiter() *Limiter {
	return &Limiter{
		limiter: rate.NewLimiter(rate.Inf, 1),
	}
}

func (l *Limiter) WithBandwith(n int64) *Limiter {
	l.limiter.SetLimit(rate.Limit(n))
	return l
}
