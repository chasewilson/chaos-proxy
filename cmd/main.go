package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/chasewilson/chaos-proxy/internal/config"
	"github.com/chasewilson/chaos-proxy/internal/logger"
	"github.com/chasewilson/chaos-proxy/internal/proxy"
)

var (
	configFile = flag.String("config", "", "path to config file")
	verbose    = flag.Bool("verbose", false, "enable verbose/debug output")
	quiet      = flag.Bool("quiet", false, "enable quite output (errors only)")
)

func main() {
	flag.Parse()
	logger.NewLogger(*verbose, *quiet)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	slog.Info("starting", "app", "chaos-proxy")
	if *configFile == "" {
		slog.Error("config file path is required",
			"flag", "-config",
			"hint", "usage: chaos-proxy -config <path-to-config.json>",
			"example", "chaos-proxy -config test-config.json")
		os.Exit(2)
	}

	slog.Info("loading config", "file", *configFile)
	routeConfigs, err := config.LoadConfig(*configFile)
	if err != nil {
		slog.Error("config validation failed",
			"file", *configFile,
			"error", err,
			"hint", "check the error messages above for specific issues and fix them in your config file")
		os.Exit(2)
	}
	slog.Info("config loaded", "file", *configFile, "routes", len(routeConfigs))
	for i, route := range routeConfigs {
		slog.Debug("route loaded",
			"index", i+1,
			"port", route.LocalPort,
			"upstream", route.Upstream,
			"dropRate", route.DropRate*100,
			"latencyMs", route.LatencyMs,
		)
	}

	slog.Info("starting listeners")

	var wg sync.WaitGroup
	for _, route := range routeConfigs {
		slog.Debug("calling ListenAndServeRoute", "port", route.LocalPort)
		wg.Add(1)
		go func(r config.RouteConfig) {
			defer wg.Done()
			listenerCtx, listenerCancel := context.WithCancel(ctx)
			defer listenerCancel()
			err := proxy.ListenAndServeRoute(listenerCtx, r)
			if err != nil {
				slog.Error("proxy listener failed",
					"port", r.LocalPort,
					"upstream", r.Upstream,
					"error", err,
					"hint", "check that the port is not already in use and you have necessary permissions")
				os.Exit(1)
			}
		}(route)
	}

	wg.Wait()
	slog.Info("all routes shut down")
}
