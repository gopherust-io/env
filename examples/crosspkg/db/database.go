package db

type Database struct {
	Host     string `env:"HOST" required:"true"`
	Password string `env:"PASSWORD" sensitive:"true"`
	Port     int    `env:"PORT" default:"5432"`
}
