#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-latest}"
count="${COUNT:-1}"
extra_args=()

while (($# > 0)); do
	case "$1" in
	--version)
		version="$2"
		shift 2
		;;
	--count)
		count="$2"
		shift 2
		;;
	*)
		extra_args+=("$1")
		shift
		;;
	esac
done

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

cd "$tmpdir"
go mod init benchrun >/dev/null
go mod edit -go=1.26.4
cat > fetch.go <<'EOF'
package benchrun

import _ "github.com/gopherust-io/env/bench"
EOF
go get -t "github.com/gopherust-io/env/bench@${version}"
rm fetch.go
if ((${#extra_args[@]} > 0)); then
	exec go test -bench=. -benchmem -count="$count" "${extra_args[@]}" github.com/gopherust-io/env/bench
fi
exec go test -bench=. -benchmem -count="$count" github.com/gopherust-io/env/bench
