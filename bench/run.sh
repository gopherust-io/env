#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

if [[ ! -f go.work ]]; then
	cp go.work.example go.work
fi

cd bench
exec go test -bench=. -benchmem -count="${COUNT:-1}" "$@"
