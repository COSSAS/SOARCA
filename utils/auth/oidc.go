package auth

import (
	"crypto/rand"
	"encoding/base64"
)

func state() (string, error) {
	rnd := make([]byte, 32)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(rnd), nil
}
