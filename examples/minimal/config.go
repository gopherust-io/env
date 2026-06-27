package minimal

//go:generate envgen -type Config -output config_env_gen.go

// Config is the smallest useful env-backed config.
type Config struct {
	Host  string `env:"HOST" default:"localhost"`
	Port  int    `env:"PORT" default:"8080"`
	Debug bool   `env:"DEBUG"`
}
