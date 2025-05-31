package jwt

import (
	"context"

	"github.com/golang-jwt/jwt"
)

type Issuer interface {
	Issue(ctx context.Context, userClaim *UserClaims) (string, error)
}

type issuer struct {
	jwtSecret string
}

func NewIssuer(jwtSecret string) Issuer {
	return &issuer{
		jwtSecret: jwtSecret,
	}
}

func (i *issuer) Issue(ctx context.Context, userClaim *UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaim)
	return token.SignedString([]byte(i.jwtSecret))
}
