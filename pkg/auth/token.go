package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
