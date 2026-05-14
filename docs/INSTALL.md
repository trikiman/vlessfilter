# Install

VlessFilter is a single Go binary. Three install paths, pick whichever:

## Option 1: Pre-built binary (fastest)

Each tagged release ships Linux + macOS binaries on GitHub Releases. Pick from <https://github.com/trikiman/vlessfilter/releases/latest>.

Linux (amd64):
```bash
curl -sSL https://github.com/trikiman/vlessfilter/releases/latest/download/vlessfilter_Linux_amd64.tar.gz \
  | tar -xz -C /tmp && sudo mv /tmp/vlessfilter /usr/local/bin/
```

## Option 2: `go install` (requires Go 1.22+)

```bash
go install github.com/trikiman/vlessfilter/cmd/vlessfilter@latest
```
Binary lands in `$GOPATH/bin` (or `$HOME/go/bin`). Make sure that's on your `$PATH`.

## Option 3: From source

```bash
git clone https://github.com/trikiman/vlessfilter.git
cd vlessfilter
go build -o bin/vlessfilter ./cmd/vlessfilter
```

## Verify it works (5 lines)

```bash
vlessfilter --help
echo "ok if you see usage above"
# Quick smoke run against the default sources (writes ./subs/ + ./README.md):
vlessfilter run --threads1 50 --threads2 5 --limit 30 --budget-min 5
ls subs/
```

## Use as a GitHub Actions workflow

The repo ships `.github/workflows/refresh.yml`, which runs the pipeline every 6 hours and commits results back to the repo. To use it on your fork:

1. Fork this repo.
2. Settings → Actions → General → Workflow permissions: **Read and write permissions**.
3. (Optional) For cross-repo pushes, add a PAT as `secrets.VLESSFILTER_TOKEN` (fine-grained, **Contents: Write** on the target repo).
4. Actions tab → "Refresh VLESS keys" → enable.

That's it. Every 6 hours the action runs, pushes fresh `subs/<CC>.txt` to your `main` branch, and the raw URLs work as static subscription links for any VLESS client.

## Use on an ephemeral VPS

See `docs/DEPLOYMENT-VPS.md` for the 2z2 Cloud Labs runbook.

## Configuration

Edit `sources.yaml` to add or remove subscription sources. The default config pulls from v2go's per-country and protocol-only outputs. See comments in the file for the schema.
