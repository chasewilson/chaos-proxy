package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/chasewilson/chaos-proxy/internal/config"
)

func main() {
	configFile := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configFile == "" {
		log.Fatal("config file path is required")
	}

	routeConfigs, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	fmt.Println(routeConfigs)
}
