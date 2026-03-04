package circuit

import (
	
	"sync"
	"time"

	"github.com/patiabhishek123/Custom-Load-Balancer/internal/balancer"
)

type Breaker struct {
	failureThreashold int 
	cooldown          time.Duration
	mu                sync.Mutex
}
//
func  NewBreaker(failureThreashold int,cooldown time.Duration) *Breaker{
	return &Breaker{
		failureThreashold: failureThreashold,
		cooldown: cooldown,
	}
}
func (br *Breaker) AllowRequest(b *balancer.Backend) bool {
    br.mu.Lock()
    defer br.mu.Unlock()

    switch b.CircuitState {

    case balancer.Open:
        if time.Since(b.LastFailureTime) > br.cooldown {
            b.SetCircuitState(balancer.HalfOpen)
            return true // allow ONE request
        }
        return false

    case balancer.HalfOpen:
        // allow only one probe request
        return false

    case balancer.Closed:
        return true
    }

    return true
}

func (br *Breaker) RecordFailures(b *balancer.Backend) {
    br.mu.Lock()
    defer br.mu.Unlock()

    if b.CircuitState == balancer.HalfOpen {
        b.SetCircuitState(balancer.Open)
        b.SetLastFailureTime(time.Now())
        return
    }

    b.IncrementFailures()

    if b.FailureCount >= br.failureThreashold {
        b.SetCircuitState(balancer.Open)
        b.SetLastFailureTime(time.Now())
    }
}
func (br *Breaker) RecordSuccess(b *balancer.Backend) {
    br.mu.Lock()
    defer br.mu.Unlock()

    b.ResetFailure()
    b.SetCircuitState(balancer.Closed)
}