package basic

import "time"

//go:generate go run ../../cmd/envgen -type Config -output config_env_gen.go

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
}

func Load() (Config, error) {
	return LoadConfig()
}
