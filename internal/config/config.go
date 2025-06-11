package config

import (
	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Database Database `envPrefix:"DATABASE_"`
	Server   Server   `envPrefix:"SERVER_"`
	Log      Log      `envPrefix:"LOG_"`
	Swagger  Swagger  `envPrefix:"SWAGGER_"`
	OnePay   OnePay   `envPrefix:"ONE_PAY_VND_"`
	SendGrip SendGrip `envPrefix:"SEND_GRIP_"`
}

var config Config

// InitConfig will load the application configuration
func init() {
	err := env.Parse(&config)
	if err != nil {
		panic("Failed to init config environment variables: " + err.Error())
	}
}

func GetConfig() Config {
	return config
}
