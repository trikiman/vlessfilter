# Ephemeral VPS Deployment (DEPRECATED)

> **This runbook is no longer usable.** It was built around 2Z2 Cloud Labs
> (`gcp.2z2.top`), which has **officially shut down** — all lab environments,
> credentials, and sandboxes were retired and will not be restored.

## Use GitHub Actions instead (recommended)

VlessFilter's recommended off-PC deployment is the **GitHub Actions** workflow
that already ships in this repo at `.github/workflows/refresh.yml`. It is
free, fully autonomous, and always-on:

- Runs every 4 hours on GitHub's own Linux runners (nothing on your PC).
- **Stage 1** does the alive/handshake check; **Stage 2** runs a real-proxy
  speed connection test on the survivors (3 passes, latency + Mbps).
- Commits fresh `subs/`, `README.md`, and `all-results.csv` back to the repo.

Setup is a one-time ~5-minute process. See **`docs/OFF-PC-DEPLOYMENT.md`
(Option A)** for the step-by-step instructions.

## Where results / speed measurements appear

- `README.md` — per-country table of **top latency (ms)** and
  **median speed (Mbps)**.
- `all-results.csv` — full raw latency + speed results for every tested key.

## Other manual fallbacks

If you specifically want a one-off manual run on a throwaway VM (no schedule),
see the other options in `docs/OFF-PC-DEPLOYMENT.md`:

- **Option B** — h2.nexus free 15-minute ephemeral VPS (no signup).
- **Option C** — Termux on an Android phone.

## Subscription URL pattern

```
https://raw.githubusercontent.com/<you>/<your-results-repo>/main/subs/<CC>.txt
```

Examples:
- `https://raw.githubusercontent.com/<you>/vlessfilter/main/subs/US.txt`
- `https://raw.githubusercontent.com/<you>/vlessfilter/main/subs/DE.txt`
