# v0.2.0

## Install

```bash
go get github.com/gopherust-io/env@v0.2.0
go install github.com/gopherust-io/env/cmd/envgen@v0.2.0
```

```go
//go:generate envgen -type Config -output config_env_gen.go
```

## What's new

### Variable expansion

```go
BaseURL string `env:"BASE_URL" default:"${NATS_URL}/api" expand`
```

### time.Time

```go
Started time.Time `env:"STARTED" layout:"2006-01-02"`
```

### .env files

```go
_ = env.LoadDotEnv(".env")
cfg, err := config.LoadConfig()
```

### Installable envgen

```bash
go install github.com/gopherust-io/env/cmd/envgen@latest
```

## CI

GitHub Actions runs `go test -race` and `go vet` on every push to `main`.

## Full changelog

[CHANGELOG.md](../CHANGELOG.md#020---2026-06-27)
