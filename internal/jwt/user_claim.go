package jwt

import (
	"github.com/golang-jwt/jwt"
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

//func NewUserClaims(user model.User, role Role) *UserClaims {
//	return &UserClaims{
//		Role:           role,
//		StandardClaims: jwt.StandardClaims{Id: strconv.Itoa(int(user.Id))},
//	}
//}
