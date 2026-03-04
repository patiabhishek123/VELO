package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/patiabhishek123/Custom-Load-Balancer/config"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/balancer"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/circuit"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/metrics"
	"github.com/patiabhishek123/Custom-Load-Balancer/internal/proxy"
)

// "github.com/patiabhishek123/Custom-Load-Balancer/server"
// "time"

func main() {
	// Command-line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	generateConfig := flag.Bool("generate-config", false, "Generate a default config file and exit")
	flag.Parse()

	// Generate default config if requested
	if *generateConfig {
		err := config.WriteDefaultConfig("config.yaml")
		if err != nil {
			log.Fatalf("Failed to generate config: %v", err)
		}
		return
	}

	// Load configuration with environment variable overrides
	cfg, err := config.LoadConfigWithEnv(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create backend pool
	pool := balancer.NewBackendPool()
	metrics.RegisterPool(pool)

	// Add backends from config
	for _, url := range cfg.Backends.URLs {
		pool.AddBackend(balancer.NewBackend(url))
		fmt.Printf("Added backend: %s\n", url)
	}

	if len(cfg.Backends.URLs) == 0 {
		log.Fatal("No backends configured")
	}

	// Create strategy based on config
	var strategy balancer.Strategy
	switch strings.ToLower(cfg.Strategy) {
	case "leastconnection":
		strategy = balancer.NewLeastCount(pool)
		fmt.Println("Using LeastConnection strategy")
	case "roundrobin":
		fallthrough
	default:
		strategy = balancer.NewRoundRobin(pool)
		fmt.Println("Using RoundRobin strategy")
	}

	// Start health check
	go pool.HealthCheck()

	// Create circuit breaker with config
	breaker := circuit.NewBreaker(cfg.Circuit.FailureThreshold, cfg.Circuit.Timeout)

	// Create load balancer
	lb := proxy.NewLoadBalancer(strategy, breaker)

	// Set up HTTP routes
	mux := http.NewServeMux()
	if cfg.Metrics.Enabled {
		mux.Handle(cfg.Metrics.Path, http.HandlerFunc(metrics.Handler))
		fmt.Printf("Metrics endpoint enabled at %s\n", cfg.Metrics.Path)
	}
	mux.Handle("/", lb)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	fmt.Printf("Starting load balancer on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
