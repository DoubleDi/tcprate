package tcprate

import (
	"context"
	"math"
	"sync"

	"golang.org/x/time/rate"
)

type Limiter struct {
	mu           sync.RWMutex
	limiter      *rate.Limiter
	limit        int
	limitPerConn int
}

func NewLimiter() *Limiter {
	return &Limiter{
		limiter:      rate.NewLimiter(rate.Inf, math.MaxInt64),
		limit:        math.MaxInt64,
		limitPerConn: math.MaxInt64,
	}
}

func (l *Limiter) WithBandwith(n int) *Limiter {
	l.mu.Lock()
	l.limiter.SetLimit(rate.Limit(n))
	l.limiter.SetBurst(n)
	l.limit = n
	l.mu.Unlock()
	return l
}

func (l *Limiter) WithPerConnBandwith(n int) *Limiter {
	l.mu.Lock()
	l.limitPerConn = n
	l.mu.Unlock()
	return l
}

func (l *Limiter) BatchAndWait(n int) (int, error) {
	r := n
	l.mu.RLock()
	if l.limit < r {
		r = l.limit
	}
	if l.limitPerConn < r {
		r = l.limitPerConn
	}
	l.mu.RUnlock()
	if err := l.limiter.WaitN(context.Background(), r); err != nil {
		return 0, err
	}
	return r, nil
}

func (l *Limiter) perConnLimit() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.limitPerConn
}
