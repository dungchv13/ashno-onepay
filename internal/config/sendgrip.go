package config

type SendGrip struct {
	ApiKey      string `env:"API_KEY" json:"api_key"`
	SenderName  string `env:"SENDER_NAME" json:"sender_name"`
	SenderEmail string `env:"SENDER_EMAIL" json:"sender_email"`
	SenderOrder string `env:"SENDER_ORDER" json:"sender_order"`
}
