package balancer

import (
	// "net/url"
	"net"
	"sync"
	"time"
)

type CircuitState int
const (
	Closed CircuitState =iota
	Open
	HalfOpen
)

type Backend struct {
	URL               string
	Healthy           bool
	ActiveConnections int64
	Weight            int
	ResponseTime      time.Duration
	mu                sync.Mutex
	CircuitState      CircuitState
	FailureCount	  int
	LastFailureTime	  time.Time
				  
}

func NewBackend(url string) *Backend{
	return &Backend{
		URL: url,
		Healthy: true,
		Weight: 1,
	}
}

func (b *Backend) IncrementConnections(){
	b.mu.Lock()
	b.ActiveConnections++
	b.mu.Unlock()
}


func (b *Backend) DecrementConnections(){
	b.mu.Lock()
	if b.ActiveConnections>0{
		b.ActiveConnections--
	}
	b.mu.Unlock()
}


func (b *Backend) IsHealthy() bool{
	b.mu.Lock()
	defer
	b.mu.Unlock()

	return b.Healthy
}

func (b *Backend) SetHealthy(state bool){
	b.mu.Lock()
	if b.Healthy!=state{
		b.Healthy=state
	}
	b.mu.Unlock()
}

func (b *Backend) SetCircuitState(state CircuitState){
	b.mu.Lock()
	defer b.mu.Unlock()
	b.CircuitState=state
}

func (b *Backend) SetLastFailureTime(time time.Time){
	b.mu.Lock()
	defer b.mu.Unlock()
	b.LastFailureTime=time
}

func (b *Backend) ResetFailure(){
	b.mu.Lock()
	defer b.mu.Unlock()

	b.FailureCount=0
}

func (b *Backend) IncrementFailures(){
	b.mu.Lock()
	defer b.mu.Unlock()
	b.FailureCount++
}

//for health check
func (b *Backend) isBackendHealthy(){
	timeout :=3*time.Second

	conn,err:=net.DialTimeout("tcp",b.URL,timeout)

	b.mu.Lock()
	defer b.mu.Unlock()

	if err!=nil{
		b.Healthy=false
		return

	}
	b.Healthy=true
	conn.Close()
}
