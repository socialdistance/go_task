package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64
	lastRefillTime time.Time
	mutex          sync.Mutex
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	duration := now.Sub(tb.lastRefillTime)
	tokensToAdd := tb.refillRate * duration.Seconds()
	tb.tokens = math.Min(tb.tokens+tokensToAdd, tb.maxTokens)
	tb.lastRefillTime = now
}

func (tb *TokenBucket) Request(tokens float64) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	tb.refill()
	if tokens <= tb.tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

func main() {
	count := 100
	ch := make(chan int, count)

	tb := NewTokenBucket(10, 1)

	wg := &sync.WaitGroup{}

	// Горутина для чтения значений из канала
	go func() {
		for value := range ch {
			fmt.Println(value)
		}
	}()

	// Горутины для записи значений в канал
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Request(1) {
				ch <- RPCCall()
			} else {
				fmt.Println("Denied:", RPCCall())
				ch <- RPCCall()
			}
		}()
		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
	close(ch)
}

func RPCCall() int {
	return rand.Int()
}
