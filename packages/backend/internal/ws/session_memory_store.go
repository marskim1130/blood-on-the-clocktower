package ws

import (
	"github.com/your-org/blood-on-the-clocktower/internal/sessionstore"
)

func newSessionMemoryStore() *sessionstore.MemoryStore { return sessionstore.NewMemoryStore() }
