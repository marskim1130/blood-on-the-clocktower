package ws

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoomRecordsRestoreCredentialSequenceAndState(t *testing.T) {
	directory := t.TempDir()
	key := []byte("0123456789abcdef0123456789abcdef")
	hub, recoveryErrors, err := NewHubWithFileRoomRecords(directory, key)
	if err != nil || len(recoveryErrors) != 0 {
		t.Fatalf("initialize: errors=%v err=%v", recoveryErrors, err)
	}
	creator := NewFakeConnection()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgCreateRoom, RequestID: "create", PlayerID: "creator", PlayerName: "Alice", MaxPlayers: 5, ScriptID: "trouble_brewing"})
	created := lastServerMessage(t, creator)
	creator.ClearMessages()
	hub.handleMessageV2(creator, ClientMessage{ProtocolVersion: 2, Type: MsgUpdateRoomSettings, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential, ClientSequence: 1, MaxPlayers: 7})
	if got := lastServerMessage(t, creator); got.RoomRevision != 2 {
		t.Fatalf("unexpected commit: %+v", got)
	}

	restored, recoveryErrors, err := NewHubWithFileRoomRecords(directory, key)
	if err != nil || len(recoveryErrors) != 0 {
		t.Fatalf("restore: errors=%v err=%v", recoveryErrors, err)
	}
	connection := NewFakeConnection()
	restored.handleMessageV2(connection, ClientMessage{ProtocolVersion: 2, Type: MsgResumeRoom, RoomID: created.RoomID, PlayerID: "creator", ResumeCredential: created.ResumeCredential})
	resumed := lastServerMessage(t, connection)
	if resumed.Code != "" || resumed.RoomRevision != 2 || resumed.NextClientSequence != 2 || resumed.State == nil || resumed.State.MaxPlayers != 7 {
		t.Fatalf("unexpected restored state: %+v", resumed)
	}

	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("room files=%v err=%v", files, err)
	}
	if info, err := os.Stat(files[0]); err != nil || info.IsDir() {
		t.Fatalf("room record missing: %v", err)
	}
}

func TestFileRoomRecordsRejectOldSnapshotFilePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := NewHubWithFileRoomRecords(path, []byte("0123456789abcdef0123456789abcdef"))
	if err == nil {
		t.Fatal("expected old single-file path to be rejected")
	}
}
