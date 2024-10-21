package auth

import (
	"kuchak/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	Email string `json:"email"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func GenerateJwtToken(email string) (string, error) {
	claims := &JwtClaims{
		email,
		false,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(config.AppConfig.SecretKey))
	if err != nil {
		return "", err
	}

	return t, nil
}
