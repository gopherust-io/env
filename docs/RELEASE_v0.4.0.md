# v0.4.0

## Install

```bash
go get github.com/gopherust-io/env@v0.4.0
go install github.com/gopherust-io/env/cmd/envgen@v0.4.0
```

## What's new

### Hot reload

```go
_ = config.ReloadConfig(&cfg)
```

### Cross-package nested types

```go
type Config struct {
    DB db.Database `prefix:"DB_"` // imported from another package
}
```

### Reflection fallback

```go
import "github.com/gopherust-io/env/reflectenv"

var cfg Config
reflectenv.Parse(&cfg)
```

## Full changelog

[CHANGELOG.md](../CHANGELOG.md#040---2026-06-27)
