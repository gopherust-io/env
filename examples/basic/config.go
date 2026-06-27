package basic

import "time"

//go:generate envgen -type Config -output config_env_gen.go

type Database struct {
	Host     string `env:"HOST" required`
	Port     int    `env:"PORT" default:"5432"`
	Password string `env:"PASSWORD" sensitive`
}

type Config struct {
	Port    int           `env:"PORT" default:"8080"`
	Debug   bool          `env:"DEBUG"`
	Timeout time.Duration `env:"TIMEOUT" default:"10s"`
	DB      Database      `prefix:"DB_"`
	Tags    []string      `env:"TAGS" sep:","`
	Labels  map[string]string `env:"LABELS" sep:"," kvsep:":"`
	Started time.Time         `env:"STARTED" layout:"2006-01-02"`
	BaseURL string            `env:"BASE_URL" default:"${NATS_URL}/api" expand`
}

func Load() (Config, error) {
	return LoadConfig()
}
