package sessionstore

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

func validateRecord(record Record) error {
	if err := validateRoomID(record.RoomID); err != nil {
		return err
	}
	if record.Revision == 0 {
		return fmt.Errorf("%w: revision must be greater than zero", ErrInvalidRecord)
	}
	if !json.Valid(record.Data) {
		return fmt.Errorf("%w: data must contain valid JSON", ErrInvalidRecord)
	}
	return nil
}

func validateRoomID(roomID string) error {
	if roomID == "" {
		return fmt.Errorf("%w: room ID is required", ErrInvalidRecord)
	}
	if roomID == "." || roomID == ".." || strings.ContainsAny(roomID, `/\\`) {
		return fmt.Errorf("%w: room ID is not path safe", ErrInvalidRecord)
	}
	for _, char := range roomID {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-' || char == '_' {
			continue
		}
		return fmt.Errorf("%w: room ID contains unsupported characters", ErrInvalidRecord)
	}
	return nil
}
