package config

type Log struct {
	Lever string `env:"LEVER" json:"lever"`
}

func (l *Log) GetLogLever() string {
	return l.Lever
}
