package model

type Items struct {
	BaseModel
	Name     string  `json:"name" gorm:"name"`
	Price    float64 `json:"price" json:"price"`
	Currency string  `json:"currency" json:"currency"`
}
