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

	routeConfigs, errs := config.LoadConfig(*configFile)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println("Config error:", err)
		}
	}

	fmt.Println(routeConfigs)
}
