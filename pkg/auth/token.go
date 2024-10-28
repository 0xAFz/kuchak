package auth

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	"github.com/rs/zerolog/log"
)

func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		log.Err(err).Msg("failed to generate random token")
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
