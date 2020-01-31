package common

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var ErrBadRateLimiter = errors.New("That rate limiter doesn't exist")
var ErrExceededRateLimit = errors.New("You're exceeding a rate limit. Please wait a while before trying again.")

// TODO: Persist rate limits to disk
type RateLimiter interface {
	LimitIP(limit, ip string) error
	LimitUser(limit string, user int) error
}

type RateData struct {
	value     int
	floorTime int
}

type RateFence struct {
	duration int
	max      int
}

// TODO: Optimise this by using something other than a string when possible
type RateLimit struct {
	data   map[string][]RateData
	fences []RateFence

	sync.RWMutex
}

func NewRateLimit(fences []RateFence) *RateLimit {
	for i, fence := range fences {
		fences[i].duration = fence.duration * 1000 * 1000 * 1000
	}
	return &RateLimit{data: make(map[string][]RateData), fences: fences}
}

func (l *RateLimit) Limit(name string, ltype int) error {
	l.Lock()
	defer l.Unlock()

	data, ok := l.data[name]
	if !ok {
		data = make([]RateData, len(l.fences))
		for i, _ := range data {
			data[i] = RateData{0, int(time.Now().Unix())}
		}
	}

	for i, field := range data {
		fence := l.fences[i]
		diff := int(time.Now().Unix()) - field.floorTime

		if diff >= fence.duration {
			field = RateData{0, int(time.Now().Unix())}
			data[i] = field
		}

		if field.value > fence.max {
			return ErrExceededRateLimit
		}

		field.value++
		data[i] = field
	}

	return nil
}

type DefaultRateLimiter struct {
	limits map[string]*RateLimit
}

func NewDefaultRateLimiter() *DefaultRateLimiter {
	return &DefaultRateLimiter{map[string]*RateLimit{
		"register": NewRateLimit([]RateFence{RateFence{int(time.Hour / 2), 1}}),
	}}
}

func (l *DefaultRateLimiter) LimitIP(limit, ip string) error {
	limiter, ok := l.limits[limit]
	if !ok {
		return ErrBadRateLimiter
	}
	return limiter.Limit(ip, 0)
}

func (l *DefaultRateLimiter) LimitUser(limit string, user int) error {
	limiter, ok := l.limits[limit]
	if !ok {
		return ErrBadRateLimiter
	}
	return limiter.Limit(strconv.Itoa(user), 1)
}
