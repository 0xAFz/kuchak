package auth

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func PasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	log.Err(err).Msg("failed to generate hash from password")
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func PasswordVerify(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
