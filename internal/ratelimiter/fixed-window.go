package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

type FixedRateWindowLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(requestsPerFrame int, timeFrame time.Duration) *FixedRateWindowLimiter {
	return &FixedRateWindowLimiter{
		clients: make(map[string]int),
		limit:   requestsPerFrame,
		window:  timeFrame,
	}
}

func (f *FixedRateWindowLimiter) Allow(ip string) (bool, time.Duration) {
	f.Lock()
	count, exists := f.clients[ip]
	f.Unlock()

	if !exists || count < f.limit {
		f.RLock()
		fmt.Println("entrou aqui")
		if !exists {
			go f.resetCount(ip)
		}

		f.clients[ip]++
		f.RUnlock()
		return true, 0
	}

	return false, f.window
}

func (f *FixedRateWindowLimiter) resetCount(ip string) {
	time.Sleep(f.window)
	f.RLock()
	delete(f.clients, ip)
	f.RUnlock()
}
