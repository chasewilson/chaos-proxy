package proxy

import (
	"errors"
	"fmt"
	"net"

	"github.com/chasewilson/chaos-proxy/internal/config"
)

func ListenAndServeRoute(route config.RouteConfig) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", route.LocalPort))
	if err != nil {
		return fmt.Errorf("error starting listener on port %d: %w", route.LocalPort, err)
	}
	defer listener.Close()

	for {
		client, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			return fmt.Errorf("error on accepting connection on port %d: %w", route.LocalPort, err)
		}

		go handleConnection(client, route.Upstream)
	}

	return nil
}

func handleConnection(client net.Conn, upstream string) error {
	defer client.Close()

	upConnection, err := net.Dial("tcp", upstream)
	if err != nil {
		return fmt.Errorf("error dialing upstream %s: %w", upstream, err)
	}
	defer upConnection.Close()
	return nil
}
