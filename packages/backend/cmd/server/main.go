package main

import (
	"log"
	"net/http"

	"github.com/your-org/blood-on-the-clocktower/internal/ws"
)

func main() {
	hub := ws.NewHub()

	http.HandleFunc("/ws", hub.HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := ":8080"
	log.Printf("WebSocket server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
