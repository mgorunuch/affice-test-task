package utils

import "sync"

func NewRateLimiter(maxRequests int) RateLimiter {
	return RateLimiter{
		currentRequests: 0,
		maxRequest:      maxRequests,
		mx:              sync.Mutex{},
	}
}

type RateLimiter struct {
	currentRequests int
	maxRequest      int

	mx sync.Mutex
}

// if connection allowed returns true
func (r *RateLimiter) RegisterConnection() bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.currentRequests >= r.maxRequest {
		return false
	}

	r.currentRequests++

	return true
}

func (r *RateLimiter) FinishConnection() {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.currentRequests > 0 {
		r.currentRequests--
	}
}
