package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type HMACCredentialCodec struct {
	key []byte
}

func NewHMACCredentialCodec(key []byte) *HMACCredentialCodec {
	return &HMACCredentialCodec{key: append([]byte(nil), key...)}
}

func (c *HMACCredentialCodec) Issue(roomID, playerID string) (string, string, error) {
	nonceBytes := make([]byte, 24)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	return nonce, c.Encode(roomID, playerID, nonce), nil
}

func (c *HMACCredentialCodec) Encode(roomID, playerID, nonce string) string {
	mac := hmac.New(sha256.New, c.key)
	mac.Write([]byte(roomID))
	mac.Write([]byte{0})
	mac.Write([]byte(playerID))
	mac.Write([]byte{0})
	mac.Write([]byte(nonce))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return nonce + "." + signature
}

func (c *HMACCredentialCodec) Verify(roomID, playerID, nonce, credential string) bool {
	parts := strings.Split(credential, ".")
	if len(parts) != 2 || parts[0] != nonce {
		return false
	}
	return hmac.Equal([]byte(c.Encode(roomID, playerID, nonce)), []byte(credential))
}
