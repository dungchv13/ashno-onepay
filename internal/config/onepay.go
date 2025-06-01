package config

type OnePay struct {
	Endpoint   string `env:"ENDPOINT" json:"endpoint"`
	MerchantID string `env:"MERCHANT_ID" json:"merchant_id"`
	AccessCode string `env:"ACCESS_CODE" json:"access-code"`
	HashCode   string `env:"HASHCODE" json:"hash_code"`
	ReturnURL  string `env:"RETURN_URL" json:"return_url"`
}
