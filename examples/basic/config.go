package basic

import "time"

//go:generate envgen -type Config -output config_env_gen.go

type Database struct {
	Host     string `env:"HOST" required:"true"`
	Password string `env:"PASSWORD" sensitive:"true"`
	Port     int    `env:"PORT" default:"5432"`
}

type Config struct {
	Started time.Time         `env:"STARTED" layout:"2006-01-02"`
	Labels  map[string]string `env:"LABELS" sep:"," kvsep:":"`
	DB      Database          `prefix:"DB_"`
	BaseURL string            `env:"BASE_URL" default:"${NATS_URL}/api" expand:"true"`
	Tags    []string          `env:"TAGS" sep:","`
	Port    int               `env:"PORT" default:"8080"`
	Timeout time.Duration     `env:"TIMEOUT" default:"10s"`
	Debug   bool              `env:"DEBUG"`
}

func Load() (Config, error) {
	return LoadConfig()
}
