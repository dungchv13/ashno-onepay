package config

type Swagger struct {
	Username string `env:"USERNAME" json:"username"`
	Password string `env:"PASSWORD" json:"password"`
	Host     string `env:"HOST" json:"host"`
	BasePath string `env:"BASE_PATH" json:"basePath"`
}

func (Swagger) GetSchemes() []string {
	return []string{"http", "https"}
}
