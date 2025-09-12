package common

import (
	"sync"
	"time"
)

type PeriodicScheduler struct {
	interval time.Duration
	callback func()
	ticker   *time.Ticker
	stopChan chan struct{}

	mu sync.Mutex
}

func NewPeriodicScheduler(interval time.Duration, cb func()) *PeriodicScheduler {
	return &PeriodicScheduler{
		interval: interval,
		callback: cb,
	}
}

func (p *PeriodicScheduler) Start() {
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

func (p *PeriodicScheduler) Stop() {
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
