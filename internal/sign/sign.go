package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

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

func Verify(msg interface{}, key, hashString string) bool {
	if key == "" {
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
