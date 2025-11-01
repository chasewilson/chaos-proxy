package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/chasewilson/chaos-proxy/internal/config"
	"github.com/chasewilson/chaos-proxy/internal/proxy"
)

func main() {
	configFile := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("config file path is required")
	}

	routeConfigs, errs := config.LoadConfig(*configFile)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println("config error:", err)
		}
	}

	for _, route := range routeConfigs {
		// possible optimization: us go routines for each connection
		err := proxy.ListenAndServeRoute(route)
		if err != nil {
			fmt.Errorf("proxy error: %w", err)
		}
	}

	fmt.Println(routeConfigs)
}
