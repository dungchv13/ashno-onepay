package model

type RegistrationOption struct {
	BaseModel
	Category string `gorm:"type:varchar(100)" json:"category"`             // e.g., "Doctor"
	Subtype  string `gorm:"type:varchar(100);default:null" json:"subtype"` // e.g., "Early-bird"

	FeeUSD float64 `gorm:"not null" json:"fee_usd"` // e.g., 500 = $500
	FeeVND int64   `gorm:"not null" json:"fee_vnd"` // e.g., 12000000 = 12,000,000 VND

	Active bool `gorm:"default:true" json:"active"`
}

type RegistrationPeriod string

const (
	EarlyBird RegistrationPeriod = "EarlyBird"
	Regular   RegistrationPeriod = "Regular"
	OnSite    RegistrationPeriod = "OnSite"
)

type RegistrationOptionFilter struct {
	Category string
	Subtype  string
}
