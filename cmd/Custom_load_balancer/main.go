package main

import (
	//"fmt"
	//"http"
	//"log"
	"fmt"
	"net/http"
	"time"

	//"github.com/patiabhishek123/Custom-Load-Balancer/internal/circuit"
	//"github.com/docker/docker/integration-cli/checker"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/balancer"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/circuit"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/proxy"
)

// "github.com/patiabhishek123/Custom-Load-Balancer/server"
// "time"

func main() {
	// go server.RunServer(5)

	// Giving servers time to start
	//     time.Sleep(100 * time.Millisecond)

	//     loadbalancer.MakeLoadBalancer(5)

	pool := balancer.NewBackendPool()
   // dynamic choosing of number of balancers
	pool.AddBackend(balancer.NewBackend("http://localhost:8081"))
	pool.AddBackend(balancer.NewBackend("http://localhost:8082"))
	pool.AddBackend(balancer.NewBackend("http://localhost:8083"))

	strategy := balancer.NewRoundRobin(pool)
	//or
	//strategy := balancer.NewLeastCount(pool)
	

	// for i := 0; i < 10; i++ {
	// 	b := strategy.NextBackend()
	// 	fmt.Println(b.URL)
	// }

	breaker :=circuit.NewBreaker(3,10*time.Second) 
	
 
	lb:=proxy.NewLoadBalancer(strategy,breaker)
	fmt.Println("Starting load balancer on :8090")
	http.ListenAndServe(":8090",lb)
	
}
