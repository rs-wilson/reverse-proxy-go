package auth

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func NewJWTAuth(sharedKey string) *JWTAuth {
	jwtAuth := &JWTAuth{
		key: sharedKey,
	}
	return jwtAuth
}

type JWTAuth struct {
	key string
}

type PomClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func (me *JWTAuth) GetToken(username string) (string, error) {
	expirationTime := time.Now().Add(10 * time.Minute) //TODO: make configurable

	claims := &PomClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// get a token with our signing type and claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign the token with our secret key
	tokenString, err := token.SignedString([]byte(me.key))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign jwt token")
	}
	return tokenString, nil
}

func (me *JWTAuth) ValidateToken(base64Token string) (string, bool) {
	// parse base64 token
	pomClaims := &PomClaims{}
	_, err := jwt.ParseWithClaims(
		base64Token,
		pomClaims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(me.key), nil
		},
	)
	if err != nil {
		log.Printf("Validate token: failed to parse jwt: %s", err.Error())
		return "", false //likely a bad token, but TODO check error types?
	}

	// Validate expiration
	if pomClaims.ExpiresAt < time.Now().Unix() {
		log.Printf("Validate token: expired token")
		return "", false // expired token!
	}

	// Valid token, so return the username
	return pomClaims.Username, true
}
