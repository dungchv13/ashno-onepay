package dto

import "ashno-onepay/internal/model"

type RegistrationRequest struct {
	RegistrationCategory string `json:"registration_category" binding:"required"`
	Nationality          string `json:"nationality"`
	DoctorateDegree      string `json:"doctorate_degree" binding:"required"`
	FirstName            string `json:"first_name"`
	MiddleName           string `json:"middle_name"`
	LastName             string `json:"last_name"`
	DateOfBirth          string `json:"date_of_birth"`
	Institution          string `json:"institution"`
	Email                string `json:"email" binding:"required,email"`
	PhoneNumber          string `json:"phone_number"`
	Sponsor              string `json:"sponsor"`

	RegistrationOption string                  `json:"registration_option" binding:"required"`
	AttendGalaDinner   bool                    `json:"attend_gala_dinner"`
	AccompanyPersons   []model.AccompanyPerson `json:"accompany_persons"`
}

type RegistrationResponse struct {
	PaymentURL string `json:"payment_url"`
	UserID     string `json:"user_id"`
}

type AccompanyPersonRegistrationRequest struct {
	Email            string                  `json:"email" binding:"required"`
	AccompanyPersons []model.AccompanyPerson `json:"accompany_persons" binding:"required,dive"`
}
