# Phase 2: Ephemeral-VPS Hardening - Research

**Researched:** 2026-05-14
**Confidence:** HIGH for kernel tuning + git auth (well-trod patterns); MEDIUM for budget enforcement (graceful xray-knife cancellation needs a defensive pattern).

## Kernel tuning (D-12)

Standard pre-flight for any high-concurrency networking on Linux:

```bash
ulimit -n 100000                                     # raise FD ceiling for current process tree
sysctl -w net.ipv4.tcp_tw_reuse=1                    # reuse TIME_WAIT sockets for outbound
sysctl -w net.ipv4.ip_local_port_range="1024 65535"  # widen ephemeral port pool
```

`ulimit` is shell-internal; from Go we use `syscall.Setrlimit(RLIMIT_NOFILE, ...)`.

`sysctl` requires root. On the 2z2 VPS the user typically has root. If not, falls back to read-only check: if current value is already permissive enough, no-op. Otherwise log warning.

Implementation: stdlib only — `syscall` for rlimit, `exec.Command("sysctl", "-w", ...)` for the sysctl writes.

## Budget enforcement (D-13)

`context.WithDeadline` propagates through all stages. xray-knife child processes started with `exec.CommandContext` are killed on context cancel.

Edge case: ungraceful kill at deadline could leave xray-knife.db locked (SQLite WAL). Mitigated by:
- xray-knife uses sqlite3 with WAL mode and 5s busy timeout
- Our selector reads with `?mode=ro&_busy_timeout=5000`

If stage 2 deadline-cancels mid-speedtest, partial results in `http_test_results` still readable.

## Git auth via http.extraheader (D-15)

Pattern (validated upstream by GitHub Actions itself):

```bash
git -c http.extraheader="Authorization: Basic $(echo -n oauth2:${TOKEN} | base64)" \
    push origin main
```

Token never written to `.git/config` (the `-c` flag is per-invocation).
Token never appears in `ps aux` (it's inside the `-c` argument, but base64-encoded — still recoverable so don't rely on this for security; rely on PAT scope).
PAT requires: `repo` scope (classic) OR `Contents: Write` (fine-grained, scoped to a single repo).

## Checkpoint cadence (D-14)

The user wants ≤2 min loss on VPS death. Implementation:

```go
ticker := time.NewTicker(2 * time.Minute)
go func() {
    for range ticker.C {
        runSelect(ctx, opts)  // re-run select stage; outputs always reflect latest DB state
        if opts.GitPush { gitCommitPush(opts) }
    }
}()
```

Cheap because `runSelect` just re-reads the SQLite DB (~10ms for 10k rows).

## Default sources expansion (D-18)

For Phase 2, sources.yaml expands to ~100 countries by listing them all under `countries:`. v2go's `Splitted-By-Country/<CC>.txt` exists for ~100 countries. Country list sourced from ISO 3166-1 alpha-2 codes that v2go publishes (US, DE, GB, FR, JP, NL, RU, etc.).

We don't enumerate every country in the default file — instead we add `countries: []` (empty list = "all countries v2go supports"). Implementation in sources package: when `countries` is empty AND a country-template source exists, fall back to a hardcoded list of common ISO codes.

## xray-knife quiet mode (D-19)

Verified: xray-knife v9.12.1 has no `--quiet` or `--no-progress` flag in the `http` command. Source confirms this (cobra command def: progressbar always on).

Workaround: pipe xray-knife stdout/stderr through a filter that drops lines containing `\r` carriage-returns (the progress-bar updater lines). Implementation in `xrayknife` wrapper using `bufio.Scanner` + custom split function.

---
*Phase: 02-vps-hardening*
*Researched: 2026-05-14*
