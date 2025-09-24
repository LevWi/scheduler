package common

import (
	"sync"
	"time"
)

type PeriodicCallback struct {
	interval time.Duration
	callback func()
	ticker   *time.Ticker
	stopChan chan struct{}

	mu sync.Mutex
}

func NewPeriodicCallback(interval time.Duration, cb func()) *PeriodicCallback {
	return &PeriodicCallback{
		interval: interval,
		callback: cb,
	}
}

func (p *PeriodicCallback) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ticker != nil {
		return
	}

	p.ticker = time.NewTicker(p.interval)
	p.stopChan = make(chan struct{})

	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.callback()
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *PeriodicCallback) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ticker == nil {
		return
	}

	p.ticker.Stop()
	close(p.stopChan)
	p.ticker = nil
	p.stopChan = nil
}
