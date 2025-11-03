package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/chasewilson/chaos-proxy/internal/chaos"
	"github.com/chasewilson/chaos-proxy/internal/config"
)

type bytesTransferred struct {
	direction string
	bytes     int64
}

func ListenAndServeRoute(ctx context.Context, route config.RouteConfig) error {
	routeLogger := slog.With("port", route.LocalPort)
	addr := fmt.Sprintf("127.0.0.1:%d", route.LocalPort)
	routeLogger.Info("starting TCP listener", "address", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		routeLogger.Error("failed to start listener", "error", err, "hint", "port may be in use or you may need elevated permissions")
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	routeLogger.Debug("listener started successfully", "address", addr)

	go func() {
		<-ctx.Done()
		routeLogger.Debug("context cancelled, closing listener", "address", addr)
		listener.Close()
	}()

	for {
		routeLogger.Debug("waiting for connection...")
		client, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				routeLogger.Debug("listener closed")
				return nil
			}

			routeLogger.Error("failed to accept connection", "error", err, "hint", "listener may have been closed unexpectedly")
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		routeLogger.Debug("connection accepted", "address", client.RemoteAddr())
		go handleConnection(ctx, client, route, routeLogger)
	}
}

func handleConnection(ctx context.Context, client net.Conn, route config.RouteConfig, routeLogger *slog.Logger) {
	defer client.Close()

	clientAddr := client.RemoteAddr().String()
	routeLogger.Debug("handling new connection", "address", clientAddr, "upstream", route.Upstream)

	server, err := net.Dial("tcp", route.Upstream)
	if err != nil {
		routeLogger.Error("failed to connect to upstream", "error", err, "hint", fmt.Sprintf("check that upstream server is running and reachable at %s", route.Upstream))
		return
	}
	defer server.Close()

	routeLogger.Info("successfully connected to upstream", "address", clientAddr, "upstream", route.Upstream)

	ritual := chaos.Ritual{
		DropRate:  route.DropRate,
		LatencyMs: route.LatencyMs,
	}
	curse := chaos.NewCurse(ritual)

	if curse.DropConnections {
		routeLogger.Info("[CHAOS] dropping connections", "address", clientAddr, "upstream", route.Upstream)
		return
	}

	go func() {
		<-ctx.Done()
		routeLogger.Debug("context cancelled, closing connection", "address", clientAddr, "upstream", route.Upstream)
		_ = client.Close()
		_ = server.Close()
	}()

	done := make(chan struct{}, 2)
	bytesResults := make(chan bytesTransferred, 2)

	routeLogger.Debug("starting data forwarding", "address", clientAddr, "upstream", route.Upstream)
	go func() {
		if curse.StartDelay > 0 {
			routeLogger.Info("[CHAOS] adding delay to upstream", "address", clientAddr, "upstream", route.Upstream, "delay", curse.StartDelay)
			select {
			case <-time.After(curse.StartDelay):

			case <-ctx.Done():
				return
			}
		}
		written, _ := io.Copy(client, server)
		bytesResults <- bytesTransferred{
			direction: "to-client",
			bytes:     written}
		done <- struct{}{}
	}()

	go func() {
		written, _ := io.Copy(server, client)
		bytesResults <- bytesTransferred{
			direction: "to-server",
			bytes:     written}
		done <- struct{}{}
	}()

	var bytesToClient, bytesToServer int64
	for i := 0; i < 2; i++ {
		result := <-bytesResults
		if result.direction == "to-client" {
			bytesToClient = result.bytes
		} else {
			bytesToServer = result.bytes
		}
	}

	totalBytes := bytesToClient + bytesToServer
	routeLogger.Info(fmt.Sprintf("bytes transferred: %d", totalBytes),
		"bytes_to_client", bytesToClient,
		"bytes_to_server", bytesToServer)

	routeLogger.Debug("connection closed", "address", clientAddr, "upstream", route.Upstream)

	<-done
}
