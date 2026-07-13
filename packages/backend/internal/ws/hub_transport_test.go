package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func handlerForHub(hub *Hub) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.HandleWebSocket)
	return mux
}

func readWireMessage(t *testing.T, connection *websocket.Conn) ServerMessage {
	t.Helper()
	_ = connection.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, payload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	var message ServerMessage
	if err := json.Unmarshal(payload, &message); err != nil {
		t.Fatalf("decode websocket message: %v", err)
	}
	return message
}

func TestWebSocketRejectsProtocolV0(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(handlerForHub(hub))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	if err := connection.WriteJSON(ClientMessage{Type: MsgCreateRoom, PlayerID: "legacy"}); err != nil {
		t.Fatal(err)
	}
	result := readWireMessage(t, connection)
	if result.Code != "UNSUPPORTED_PROTOCOL" {
		t.Fatalf("protocol v0 must be rejected: %+v", result)
	}
}

func TestWebSocketAcceptsProtocolV2(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(handlerForHub(hub))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	request := ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "wire-create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5}
	if err := connection.WriteJSON(request); err != nil {
		t.Fatal(err)
	}
	result := readWireMessage(t, connection)
	if result.Type != "CREATE_ROOM_RESULT" || result.RoomID == "" || result.ResumeCredential == "" {
		t.Fatalf("unexpected create result: %+v", result)
	}
}
