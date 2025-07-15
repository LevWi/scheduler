package common

import (
	"sync"

	"golang.org/x/time/rate"
)

type LimitsTable[Key comparable] struct {
	m        map[Key]*rate.Limiter
	mu       sync.Mutex
	settings RequestLimitSettings
}

type RequestLimitSettings struct {
	Limit rate.Limit
	Burst int
}

func (l *LimitsTable[Key]) Allow(k Key) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.m[k]
	if !exists {
		limiter = rate.NewLimiter(l.settings.Limit, l.settings.Burst)
		l.m[k] = limiter
	}
	//TODO add info about needed delay?
	return limiter.Allow()
}

func (l *LimitsTable[Key]) SetNewSettings(s RequestLimitSettings) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, limiter := range l.m {
		limiter.SetLimit(s.Limit)
		limiter.SetBurst(s.Burst)
	}
	l.settings = s
}

func NewRestrictionList[Key comparable](s RequestLimitSettings) *LimitsTable[Key] {
	return &LimitsTable[Key]{settings: s}
}
