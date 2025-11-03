package testserver

import (
	"fmt"
	"log/slog"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test server response at %s\n", r.Host)
	fmt.Fprintf(w, "Under heavy load (of self-doubt and coffee).\n")
}

func NewTestServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)

	slog.Info("starting test HTTP server", "address", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("test server failed", "address", addr, "error", err)
	}
}
