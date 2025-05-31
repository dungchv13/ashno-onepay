package jwt

import (
	"ashno-onepay/internal/errors"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt"
	"time"
)

type Validator interface {
	Validate(ctx context.Context, token string) (*UserClaims, error)
}

type validatorImpl struct {
	jwtSecret string
}

var (
	ErrInvalidToken   = fmt.Errorf("invalid token")
	ErrorExpiredToken = fmt.Errorf("token expired")
	ErrInvalidSession = fmt.Errorf("invalid session")
)

func NewValidator(jwtSecret string) Validator {
	return &validatorImpl{
		jwtSecret: jwtSecret,
	}
}

func (v *validatorImpl) getClaim(jwtToken string) (*UserClaims, error) {
	claims := new(UserClaims)
	token, err := jwt.ParseWithClaims(jwtToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(v.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*UserClaims), nil
}

func (v *validatorImpl) Validate(ctx context.Context, jwtToken string) (*UserClaims, error) {
	claims, err := v.getClaim(jwtToken)
	if err != nil {
		return nil, errors.ErrInvalidSession.Wrap(err)
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return claims, errors.ErrInvalidSession.Reform(ErrorExpiredToken.Error())
	}
	return claims, err
}
