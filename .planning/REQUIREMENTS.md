# Requirements: VlessFilter v1.2 + v1.3

## Milestone v1.2 — Liveness Validation

The only question this milestone answers: is this VLESS key actually
working *right now* from the user's network?

### Active

- [ ] **LIVE-01**: When a config passes stage-1 handshake, it must be
      retested 2 more times (3 total attempts) before being marked alive.
      Configs with 3/3 success → "alive". 1-2/3 success → "flaky". 0/3 → "dead".
      Only "alive" configs are eligible for `subs/<CC>.txt` publication.
- [ ] **LIVE-02**: Each test attempt must complete a real HTTP request
      through the proxy (not just TLS handshake). Use xray-knife --speedtest
      mode which already does this.
- [ ] **LIVE-03**: A config previously marked alive that fails 1 of N
      retests is NOT immediately demoted to flaky. Sticky-alive logic:
      requires 2 consecutive failures across runs before demoting. Prevents
      the "9 → 12 → 9" oscillation users saw.
- [ ] **LIVE-04**: New `--profile dev` flag runs the entire pipeline against
      a 500-config subset in under 2 minutes wall-clock. Used for verifying
      filter logic changes without waiting for a 6-hour scheduled run.

## Milestone v1.3 — Country Identification

The only question this milestone answers: for a verified-alive config,
where does it actually exit?

### Active

- [ ] **GEO-01**: Test URL changed from cloudflare.com/cdn-cgi/trace
      (which only returns Cloudflare PoP location) to ipinfo.io/json
      (which returns the actual public IP of the proxy's egress point).
      Parse `country` field from response.
- [ ] **GEO-02**: Country label is "stable" only when 2+ test attempts
      agree on the same country code. Single-test results are tagged as
      `unverified` and held back from publication until corroborated.
- [ ] **GEO-03**: Configs with 2+ different exit countries across history
      go to `subs/rotating.txt` with `🌐 ROTATING` remark, not `subs/<CC>.txt`.
      (Already shipped in v1.1; verify still working with v1.3 stricter
      consensus rule.)
- [ ] **GEO-04**: Post-publication accuracy probe: after `subs/<CC>.txt`
      files are written, the pipeline samples up to 5 random keys per
      country, routes a real HTTP request through each to ipinfo.io,
      compares actual exit to the published label. If accuracy across all
      sampled countries is <80%, the run logs an ERROR and refuses to push
      the bad output (keeps the previous-run output live instead).

### Out of Scope (v1.2 + v1.3)

- Distributed validation from multiple geographic perspectives (would need
  CF Workers + a second free VPS — explored, deferred to v1.4 if needed)
- Replacing xray-knife with custom proxy engine (still good enough for
  these milestones)
- Telegram-channel scraping (already covered by upstream sources)

## Future Requirements (post-v1.3)

- **DIST-01**: Multi-perspective country validation by running probes from
  Cloudflare Workers (200+ edge locations) — recovers configs that fail
  from user's home network but work elsewhere
- **TIER-01**: Tier-2 fallback for countries with zero stable-alive keys
  (publish flaky-but-recently-passed configs labeled as "best effort")
- **API-01**: HTTP API endpoint (`/api/v1/country/{cc}`) that returns top-3
  for that country as JSON, with stability score and last-tested time

## Traceability

| REQ | Phase |
|-----|-------|
| LIVE-01 | 4 |
| LIVE-02 | 4 |
| LIVE-03 | 4 |
| LIVE-04 | 5 |
| GEO-01 | 6 |
| GEO-02 | 6 |
| GEO-03 | 6 (verify) |
| GEO-04 | 7 |
