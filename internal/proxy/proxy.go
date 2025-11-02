package proxy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/chasewilson/chaos-proxy/internal/chaos"
	"github.com/chasewilson/chaos-proxy/internal/config"
)

func ListenAndServeRoute(route config.RouteConfig) error {
	routeLogger := slog.With("port", route.LocalPort)
	addr := fmt.Sprintf("127.0.0.1:%d", route.LocalPort)
	routeLogger.Info("starting TCP listener", "address", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		routeLogger.Error("failed to start listener",
			"error", err,
			"hint", "port may be in use or you may need elevated permissions")
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	routeLogger.Debug("listener started successfully", "address", addr)

	for {
		routeLogger.Info("waiting for connection...")
		client, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				routeLogger.Info("listener closed")
				return nil
			}

			routeLogger.Error("failed to accept connection",
				"error", err,
				"hint", "listener may have been closed unexpectedly")
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		routeLogger.Info("connection accepted", "address", client.RemoteAddr())
		go handleConnection(client, route, routeLogger)
	}
}

func handleConnection(client net.Conn, route config.RouteConfig, routeLogger *slog.Logger) {
	defer client.Close()

	clientAddr := client.RemoteAddr().String()
	routeLogger.Debug("handling new connection", "address", clientAddr, "upstream", route.Upstream)

	server, err := net.Dial("tcp", route.Upstream)
	if err != nil {
		routeLogger.Error("failed to connect to upstream",
			"error", err,
			"hint", fmt.Sprintf("check that upstream server is running and reachable at %s", route.Upstream))
		return
	}
	defer server.Close()

	routeLogger.Info("successfully connected to upstream",
		"address", clientAddr,
		"upstream", route.Upstream)

	ritual := chaos.Ritual{
		DropRate:  route.DropRate,
		LatencyMs: route.LatencyMs,
	}
	curse := chaos.NewCurse(ritual)

	if curse.DropConnections {
		routeLogger.Error("[CHAOS] dropping connections",
			"address", clientAddr,
			"upstream", route.Upstream,
			"error", err)
		client.Close()
		server.Close()
		return
	}

	done := make(chan struct{}, 2)
	go func() {
		if curse.StartDelay > 0 {
			routeLogger.Info("[CHAOS] adding delay to upstream",
				"address", clientAddr,
				"upstream", route.Upstream,
				"delay", curse.StartDelay)
			time.Sleep(curse.StartDelay)
		}
		io.Copy(client, server)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(server, client)
		done <- struct{}{}
	}()

	<-done
	// Requirement is that it waits for either side to close. Can remove comment if needed.
	// <-done
}
