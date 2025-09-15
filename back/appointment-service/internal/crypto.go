package common

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSecretKey(length int) string {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}

	return base64.RawURLEncoding.EncodeToString(key)
}
