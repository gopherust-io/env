package crosspkg

import "github.com/gopherust-io/env/examples/crosspkg/db"

//go:generate envgen -type Config -output config_env_gen.go

type Config struct {
	DB   db.Database `prefix:"DB_"`
	Port int         `env:"PORT" default:"8080"`
}
