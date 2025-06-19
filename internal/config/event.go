package config

type Event struct {
	Name  string `env:"NAME" json:"name"`
	Date  string `env:"DATE" json:"date"`
	Venue string `env:"VENUE" json:"venue"`
}
