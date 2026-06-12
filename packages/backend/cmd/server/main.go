package main

import (
	"log"
	"net/http"
	"os"

	"github.com/your-org/blood-on-the-clocktower/internal/ws"
)

func main() {
	hub := ws.NewHub()
	if redisURL := os.Getenv("CLOCKTOWER_REDIS_URL"); redisURL != "" {
		persistentHub, err := ws.NewHubWithRedisSnapshot(redisURL, os.Getenv("CLOCKTOWER_REDIS_KEY"))
		if err != nil {
			log.Fatalf("failed to restore Redis snapshot: %v", err)
		}
		hub = persistentHub
		log.Printf("Redis snapshot persistence enabled")
	} else if snapshotPath := os.Getenv("CLOCKTOWER_SNAPSHOT_PATH"); snapshotPath != "" {
		persistentHub, err := ws.NewHubWithFileSnapshot(snapshotPath)
		if err != nil {
			log.Fatalf("failed to restore snapshot %s: %v", snapshotPath, err)
		}
		hub = persistentHub
		log.Printf("game snapshot persistence enabled at %s", snapshotPath)
	}

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
