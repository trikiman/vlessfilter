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
#
# sources.txt IS DOCUMENTATION ONLY. Nothing reads it: the pipeline loads
# sources.yaml via internal/sources.Load (pipeline.go sets SourcesPath to
# "./sources.yaml"). No workflow runs this script either, so sources.txt only
# changes when a human runs it — it can silently drift from sources.yaml.
# Re-run this after editing sources.yaml if you want the manifest to match.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ ! -f bin/vlessfilter ]; then
  echo "bin/vlessfilter not found; building..."
  go build -o bin/vlessfilter ./cmd/vlessfilter
fi

./bin/vlessfilter sources-list --sources sources.yaml > sources.txt
echo "wrote $(wc -l < sources.txt) URLs to sources.txt"
