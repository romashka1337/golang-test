package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("veryhardpass")

type Cred struct {
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(name string) string {
	cred := &Cred{
		UserId: name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, cred).
		SignedString(secret)
	if err != nil {
		log.Printf("%v", err)
	}
	return token
}

func ValidateToken(str string) (*Cred, error) {
	cred := &Cred{}
	_, err := jwt.ParseWithClaims(
		str,
		cred,
		func(token *jwt.Token) (any, error) { return secret, nil },
	)
	return cred, err
}
