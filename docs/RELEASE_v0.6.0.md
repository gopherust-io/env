# v0.6.0

## Install

```bash
go get github.com/gopherust-io/env@v0.6.0
go install github.com/gopherust-io/env/cmd/envgen@v0.6.0
```

Regenerate loaders after upgrading:

```bash
go generate ./...
```

## What's new

### Config Doctor

Audit a snapshot against the same schema as codegen — typos, orphans, silent defaults, missing required, unmarked secrets:

```bash
envgen doctor -type Config -env-file .env -mode prod
```

```go
report := config.AuditConfigFromEnviron()
for _, f := range report.Findings {
    log.Println(f)
}
```

### Hardening

- Sensitive values stay out of parse errors and audit default literals
- `ReloadConfig` clears fields when env keys are removed
- Safer dotenv parsing and `envgen -output` path containment

## Full changelog

[CHANGELOG.md](../CHANGELOG.md#060---2026-08-02)
