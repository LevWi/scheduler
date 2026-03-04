package common

import (
	"sync"

	"golang.org/x/time/rate"
)

// TODO remove not needed keys function
type LimitsTable[Key comparable] struct {
	m        map[Key]*rate.Limiter
	mu       sync.Mutex
	producer RequestLimitUpdate
}

type RequestLimitUpdate interface {
	Update(*rate.Limiter) *rate.Limiter
}

type RequestLimitUpdateFunc func(*rate.Limiter) *rate.Limiter

func (fn RequestLimitUpdateFunc) Update(in *rate.Limiter) *rate.Limiter {
	return fn(in)
}

func (l *LimitsTable[Key]) Allow(k Key) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.m[k]
	if !exists {
		limiter = l.producer.Update(nil)
		l.m[k] = limiter
	}
	//TODO add info about needed delay?
	return limiter.Allow()
}

func (l *LimitsTable[Key]) SetProducer(p RequestLimitUpdate) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, limiter := range l.m {
		l.producer.Update(limiter)
	}
	l.producer = p
}

func NewLimitsTable[Key comparable](p RequestLimitUpdate) *LimitsTable[Key] {
	return &LimitsTable[Key]{
		m:        make(map[Key]*rate.Limiter),
		producer: p}
}
