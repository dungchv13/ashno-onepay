package jwt

import (
	"ashno-onepay/internal/model"
	"github.com/golang-jwt/jwt"
	"strconv"
)

type UserClaims struct {
	jwt.StandardClaims
	Role Role `json:"role"`
}

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

func NewUserClaims(user model.User, role Role) *UserClaims {
	return &UserClaims{
		Role:           role,
		StandardClaims: jwt.StandardClaims{Id: strconv.Itoa(int(user.Id))},
	}
}
