package main

import (
	"log"
	"net/http"
	"os"

	"github.com/your-org/blood-on-the-clocktower/internal/ws"
)

func main() {
	credentialKey := []byte(os.Getenv("CLOCKTOWER_CREDENTIAL_KEY"))
	var hub *ws.Hub
	if redisURL := os.Getenv("CLOCKTOWER_REDIS_URL"); redisURL != "" {
		persistentHub, recoveryErrors, err := ws.NewHubWithRedisRoomRecords(redisURL, os.Getenv("CLOCKTOWER_REDIS_KEY"), credentialKey)
		if err != nil {
			log.Fatalf("failed to restore Redis room records: %v", err)
		}
		hub = persistentHub
		for _, recoveryErr := range recoveryErrors {
			log.Printf("skipped room record: %v", recoveryErr)
		}
		log.Printf("Redis room record persistence enabled; single active backend instance required")
	} else if snapshotPath := os.Getenv("CLOCKTOWER_SNAPSHOT_PATH"); snapshotPath != "" {
		persistentHub, recoveryErrors, err := ws.NewHubWithFileRoomRecords(snapshotPath, credentialKey)
		if err != nil {
			log.Fatalf("failed to restore room records from %s: %v", snapshotPath, err)
		}
		hub = persistentHub
		for _, recoveryErr := range recoveryErrors {
			log.Printf("skipped room record: %v", recoveryErr)
		}
		log.Printf("file room record persistence enabled at %s; single active backend instance required", snapshotPath)
	} else {
		ephemeralHub, err := ws.NewEphemeralProductionHub(credentialKey)
		if err != nil {
			log.Fatalf("failed to initialize credential signing: %v", err)
		}
		hub = ephemeralHub
		log.Printf("in-memory room storage enabled; data will not survive restart; single active backend instance required")
	}

	http.HandleFunc("/ws", hub.HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !hub.Healthy() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("persistence conflict"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8080"
	log.Printf("WebSocket server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
