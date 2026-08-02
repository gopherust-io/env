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

Configured for Scorecard Branch-Protection / Code-Review:

1. **Branch protection on `main`** — require 1 approving review, dismiss stale reviews, require last-push approval, require `test` + `lint`, linear history, no force-push/deletion, enforce for admins
2. **Ruleset `main`** — targets `refs/heads/main`; blocks deletion / non-FF; required status checks `test` + `lint`
3. **Code scanning** — Scorecard SARIF + CodeQL default setup
4. Re-run — **Actions → Scorecard → Run workflow** after settings or dependency changes

## Alert remediation notes

| Check | How we address it |
|-------|-------------------|
| Branch-Protection | Branch protection + ruleset above |
| Code-Review | Same; merge only via reviewed PRs (collaborator approval). Score improves as recent commits are approved PRs |
| Vulnerabilities | Keep `bench` / deps on OSV-clean versions (e.g. `golang.org/x/text` ≥ 0.39.0) |
| Maintained | Scorecard requires repo age **> 90 days** (created 2026-06-27). Keep weekly activity; reassess after ~2026-09-25 |
| CII-Best-Practices | Register at [bestpractices.dev](https://www.bestpractices.dev/en/projects/new?repo_url=https%3A%2F%2Fgithub.com%2Fgopherust-io%2Fenv) (manual login). Even an in-progress badge improves the score |

## Related hardening in-repo

- [`SECURITY.md`](../SECURITY.md) — vulnerability reporting (Security-Policy check)
- [`.github/dependabot.yml`](../.github/dependabot.yml) — dependency update tool
- SHA-pinned Actions in CI and Scorecard workflows (Pinned-Dependencies)
