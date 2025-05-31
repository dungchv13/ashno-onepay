package dto

import (
	"ashno-onepay/internal/jwt"
)

type LoginRequest struct {
	Email    string `json:"email" gorm:"email"`
	Password string `json:"password" gorm:"password"`
}

type LoginResponse struct {
	Id    int64    `json:"id"`
	Role  jwt.Role `json:"role"`
	Token string   `json:"token"`
}

type RegisterRequest struct {
	Email       string `json:"email" required:"true"`
	Password    string `json:"password" required:"true"`
	Phone       string `json:"phone"`
	Nationality string `json:"nationality"`
	FirstName   string `json:"first_name"`
	MiddleName  string `json:"middle_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Institution string `json:"institution"`
	SponsorBy   string `json:"sponsor_by"`
}
