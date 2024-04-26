package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GetHmacString(body []byte, key string) (string, error) {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func Verify(msg []byte, key, hashString string) (bool, error) {
	sig, err := hex.DecodeString(hashString)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(key))
	_, err = mac.Write(msg)
	if err != nil {
		return false, err
	}

	return hmac.Equal(sig, mac.Sum(nil)), nil
}
