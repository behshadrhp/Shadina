package auth

import "github.com/golang-jwt/jwt/v5"

type JWTAuthenticator struct {
	secretKey string
	iss       string
	aud       string
}

func NewJWTAuthenticator(secretKey, iss, aud string) JWTAuthenticator {
	return JWTAuthenticator{
		secretKey,
		iss,
		aud,
	}
}

func (ja JWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(ja.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (ja JWTAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return nil, nil
}
