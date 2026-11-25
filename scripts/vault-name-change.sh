#!/usr/bin/env bash
# Runs the vault program against the second test ground and tracks what the vault folder holds after
# each run. The second run happens after the source changes the node part of an identifier, so the
# different name is produced by the mapping rather than by hand. The source is restored on exit.
set -euo pipefail

module="/home/galdoba/go/src/github.com/Galdoba/go-graph"
ground="$module/testdata/second/second.go"
vault="/home/galdoba/Documents/Obsidian/GoRepository/github.com/Galdoba/go-graph/testdata/second"

before="func New() Thing { return Thing{} }"
after="func NewThing() Thing { return Thing{} }"

restore() {
  sed -i "s|$after|$before|" "$ground"
  echo "restored: $before"
}
trap restore EXIT

state() {
  echo "=== $1 ==="
  if [ -d "$vault" ]; then
    ls -1 "$vault" | sort
  else
    echo "no folder yet: $vault"
  fi
}

state "vault before the first run"
(cd "$module" && go run ./cmd/graphnodes ./testdata/second)

state "vault after the first run"
sed -i "s|$before|$after|" "$ground"

(cd "$module" && go run ./cmd/graphnodes ./testdata/second)
state "vault after the second run"
