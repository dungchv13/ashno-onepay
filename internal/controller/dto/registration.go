package dto

type RegistrationResponse struct {
	PaymentURL string `json:"payment_url"`
	UserID     string `json:"user_id"`
}
