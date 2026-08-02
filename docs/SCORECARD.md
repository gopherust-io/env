# OpenSSF Scorecard

This repository runs [OpenSSF Scorecard](https://scorecard.dev/) via GitHub Actions.

| Item | Location |
|------|----------|
| Workflow | [`.github/workflows/scorecard.yml`](../.github/workflows/scorecard.yml) |
| Badge | README (updates after first successful publish) |
| Results API | `https://api.scorecard.dev/projects/github.com/gopherust-io/env` |
| Viewer | [scorecard.dev viewer](https://scorecard.dev/viewer/?uri=github.com/gopherust-io/env) |

## What the workflow does

1. Runs on pushes to `main`, weekly schedule, and manual `workflow_dispatch`
2. Writes SARIF results and uploads them to **GitHub Code Scanning**
3. Publishes results to the OpenSSF API (`publish_results: true`) for the public badge

## Repo settings (maintainers)

After merging the workflow:

1. **Actions** — ensure Actions are enabled for the repository
2. **Code scanning** — enable the default Code Scanning / code security features so SARIF uploads succeed (GitHub → Settings → Code security)
3. **Branch protection** (recommended for higher scores) — require PR reviews and status checks on `main`
4. First run — open **Actions → Scorecard → Run workflow**, or push to `main`

Branch-protection checks sometimes need a PAT with `admin:repo_hook` / fine-grained admin read; public repos usually get useful scores with the default `GITHUB_TOKEN`.

## Related hardening in-repo

- [`SECURITY.md`](../SECURITY.md) — vulnerability reporting (Security-Policy check)
- [`.github/dependabot.yml`](../.github/dependabot.yml) — dependency update tool
- SHA-pinned Actions in CI and Scorecard workflows (Pinned-Dependencies)
