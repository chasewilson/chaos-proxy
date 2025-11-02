package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/chasewilson/chaos-proxy/internal/config"
	"github.com/chasewilson/chaos-proxy/internal/proxy"
)

func main() {
	fmt.Println("==> Starting chaos-proxy...")

	configFile := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("config file path is required")
	}

	fmt.Printf("==> Loading config from: %s\n", *configFile)
	routeConfigs, errs := config.LoadConfig(*configFile)
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println("config error:", err)
		}
		log.Fatal("config validation failed")
	}

	fmt.Printf("==> Loaded %d route(s)\n", len(routeConfigs))
	for i, route := range routeConfigs {
		fmt.Printf("    Route %d: localhost:%d -> %s (drop: %.1f%%, latency: %dms)\n",
			i+1, route.LocalPort, route.Upstream, route.DropRate*100, route.LatencyMs)
	}

	fmt.Println("==> Starting listeners...")
	var wg sync.WaitGroup
	for _, route := range routeConfigs {
		// possible optimization: us go routines for each connection
		fmt.Printf("==> Calling ListenAndServeRoute for port %d...\n", route.LocalPort)
		wg.Go(func() {
			err := proxy.ListenAndServeRoute(route)
			if err != nil {
				log.Fatalf("proxy error on port %d: %v", route.LocalPort, err)
			}
		})
	}

	wg.Wait()
	fmt.Println("==> All routes shut down")
}
