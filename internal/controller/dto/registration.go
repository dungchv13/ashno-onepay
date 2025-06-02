package dto

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

	RegistrationOption string `gorm:"foreignKey:RegistrationOptionID"`
	AttendGalaDinner   bool   `json:"attend_gala_dinner"`
}

type RegistrationResponse struct {
	PaymentURL string `json:"payment_url"`
	UserID     string `json:"user_id"`
}
