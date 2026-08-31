package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	window time.Time
	count  int
}
type Limiter struct {
	mu               sync.Mutex
	perSecond, burst int
	users            map[string]entry
}

func New(perSecond, burst int) *Limiter {
	return &Limiter{perSecond: perSecond, burst: burst, users: make(map[string]entry)}
}
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	e := l.users[key]
	if e.window.IsZero() || now.Sub(e.window) >= time.Second {
		e = entry{window: now}
	}
	limit := l.perSecond
	if l.burst > limit {
		limit = l.burst
	}
	if e.count >= limit {
		return false
	}
	e.count++
	l.users[key] = e
	return true
}
