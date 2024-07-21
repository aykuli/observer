// Package sign provides methods to get cryptographic hash and verifying string by the key
// based on  Keyed-Hash Message Authentication Code (HMAC) provided by hmac package.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// GetHmacString encodes bytes and returns string by provided key
func GetHmacString(body []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Verify returns true/false value based on provided key
func Verify(msg interface{}, key, hashString string) bool {
	if key == "" || hashString == "" {
		return true
	}

	byteData, err := json.Marshal(msg)
	if err != nil {
		return false
	}

	sig, err := hex.DecodeString(hashString)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(key))
	if _, err = mac.Write(byteData); err != nil {
		return false
	}

	return hmac.Equal(sig, mac.Sum(nil))
}
