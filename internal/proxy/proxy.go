package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/chasewilson/chaos-proxy/internal/config"
)

func ListenAndServeRoute(route config.RouteConfig) error {
	addr := fmt.Sprintf("127.0.0.1:%d", route.LocalPort)
	fmt.Printf("    [Port %d] Starting TCP listener on %s...\n", route.LocalPort, addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("error starting listener on port %d: %w", route.LocalPort, err)
	}
	defer listener.Close()

	fmt.Printf("    [Port %d] Listener started successfully, waiting for connections...\n", route.LocalPort)

	for {
		fmt.Printf("    [Port %d] Calling Accept() - BLOCKING until connection...\n", route.LocalPort)
		client, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Printf("    [Port %d] Listener closed\n", route.LocalPort)
				return nil
			}

			return fmt.Errorf("error on accepting connection on port %d: %w", route.LocalPort, err)
		}

		fmt.Printf("    [Port %d] Connection accepted from %s\n", route.LocalPort, client.RemoteAddr())
		go handleConnection(client, route)
	}
}

func handleConnection(client net.Conn, route config.RouteConfig) {
	defer client.Close()

	clientAddr := client.RemoteAddr().String()
	fmt.Printf("        [Connection %s] Handling new connection, dialing upstream %s...\n", clientAddr, route.Upstream)

	server, err := net.Dial("tcp", route.Upstream)
	if err != nil {
		fmt.Printf("        [Connection %s] ERROR: Failed to connect to upstream %s: %v\n", clientAddr, route.Upstream, err)
		return
	}
	defer server.Close()

	fmt.Printf("        [Connection %s] Successfully connected to upstream %s\n", clientAddr, route.Upstream)
	done := make(chan struct{}, 2)
	go func() {
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
