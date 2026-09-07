package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marskim1130/blood-on-the-clocktower/internal/ws"
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

	cleanupContext, stopCleanup := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopCleanup()
	go hub.RunInactiveRoomCleanup(cleanupContext)
	log.Printf("room cleanup enabled: only successful persisted writes refresh the seven-day inactivity window")

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
	server := &http.Server{Addr: addr, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-cleanupContext.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()
	log.Printf("WebSocket server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to serve: %v", err)
	}
}
