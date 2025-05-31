package config

import (
	"github.com/pkg/errors"
	"time"
)

type Server struct {
	Host        string `env:"HOST" json:"host"`
	Port        string `env:"PORT" json:"port"`
	ReadTimeout string `env:"READ_TIMEOUT" json:"readTimeout"`
	JwtKey      string `env:"JWT_KEY" json:"jwtKey"`
	EncryptKey  string `env:"ENCRYPT_KEY" json:"encryptKey"`
}

func (s Server) GetAddr() string {
	return s.Host + ":" + s.Port
}

func (s Server) GetReadTimeout() time.Duration {
	duration, err := time.ParseDuration(s.ReadTimeout)
	if err != nil {
		panic(errors.Wrap(err, "Failed to parse read timeout"))
	}
	return duration
}
