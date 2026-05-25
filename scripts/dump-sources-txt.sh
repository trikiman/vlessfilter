#!/usr/bin/env bash
# Regenerate sources.txt from sources.yaml.
#
# Usage:   bash scripts/dump-sources-txt.sh
# Output:  sources.txt at repo root, one expanded URL per line.
#
# Why: gives a stable, diffable manifest of every subscription URL the
# pipeline pulls from. Useful for:
#   - reviewing source-set changes in PRs
#   - manually probing a single source for breakage
#   - bootstrapping a new pipeline machine without parsing YAML
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f bin/vlessfilter ]; then
  echo "bin/vlessfilter not found; building..."
  go build -o bin/vlessfilter ./cmd/vlessfilter
fi

./bin/vlessfilter sources-list --sources sources.yaml > sources.txt
echo "wrote $(wc -l < sources.txt) URLs to sources.txt"
