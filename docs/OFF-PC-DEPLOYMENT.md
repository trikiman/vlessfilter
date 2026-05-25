# Off-PC Deployment

Three ways to run the VlessFilter pipeline without using your own computer.
All options are **free** and require no recurring billing.

## Option A: GitHub Actions on a NEW account (recommended)

**Cost:** $0 — public repos get unlimited Actions minutes on the free tier.
**Setup time:** ~5 minutes.
**Always-on:** Yes (cron schedule on GitHub's runners).

### Steps

1. **Create a new GitHub account** at https://github.com/signup. Use any
   throwaway email. No card needed for free-tier Actions on public repos.
2. **Fork** `https://github.com/trikiman/vlessfilter` to your new account.
3. **Generate a PAT** on the upstream account (`trikiman` — the account that
   should receive results):
   - Go to https://github.com/settings/personal-access-tokens/new
   - Repository access: only `trikiman/vlessfilter`
   - Permissions: Contents = Read and write
   - Copy the generated token (starts with `github_pat_...`)
4. **In the FORKED repo**, go to Settings → Secrets and variables → Actions:
   - New repository secret: `PUSH_TOKEN` = the PAT from step 3
5. **Enable Actions** in the fork: Settings → Actions → General → Allow all.
6. **Trigger first run**: Actions tab → "refresh" workflow → Run workflow.

The `.github/workflows/refresh.yml` file already exists in the repo and runs
every 6 hours. Results are committed to the upstream
`trikiman/vlessfilter` main branch using the PAT.

### Pros
- Fully automated
- No card/billing
- GitHub's infrastructure handles the runtime

### Cons
- Requires creating a new GitHub account (the original `trikiman` one is
  blocked from Actions due to a $0.30 unpaid Copilot Premium overage)
- Logs visible to anyone with read access to the fork

---

## Option B: Free-tier always-on Linux VM (Oracle Cloud / similar)

**Cost:** $0 — Oracle Cloud Always-Free tier (4 ARM cores, 24GB RAM, runs
forever, NOT a 30-day trial).
**Setup time:** ~30 minutes.
**Always-on:** Yes (cron on the VM).

### Steps

1. **Sign up at Oracle Cloud**: https://www.oracle.com/cloud/free/
   - Card verification required, but no charge for Always-Free resources.
   - May fail for some Russian cards — try alternative tier or fall back
     to Google Cloud Free Tier or AWS Free Tier.
2. **Provision an Ampere ARM VM**:
   - Compute → Instances → Create Instance
   - Shape: VM.Standard.A1.Flex (ARM)
   - 4 OCPU, 24 GB RAM (max free quota)
   - Image: Canonical Ubuntu 24.04
   - Add SSH public key
3. **Open SSH** to the VM (use the public IP shown in console):
   ```
   ssh ubuntu@<public-ip>
   ```
4. **Generate a PAT** as in Option A step 3.
5. **Run the install script** on the VM:
   ```bash
   curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/install-always-on.sh \
     | bash -s -- github_pat_xxx
   ```
6. **Verify** with:
   ```bash
   crontab -l           # should show refresh.sh at 0,6,12,18 * * *
   $HOME/.vlessfilter/refresh.sh &   # trigger first run manually
   tail -f $(ls -1t $HOME/.vlessfilter/refresh-*.log | head -1)
   ```

### Pros
- Truly always-on, no monthly minute caps
- Faster CPUs than GitHub runners
- Independent of GitHub billing status

### Cons
- One-time card verification (Oracle, Google, AWS all require this)
- ~30 min setup time
- Russian customers may face account-creation friction

---

## Option C: h2.nexus 15-minute ephemeral VPS (manual trigger, no signup)

**Cost:** $0 — h2.nexus gives free 15-min VPS (4 CPU, 8GB RAM, 1Gbps) with no
account.
**Setup time:** Zero.
**Always-on:** No — manual trigger when you want fresh keys (recommend
2-4× per day).

### Steps

1. **Generate a PAT** as in Option A step 3 (one-time). Save it somewhere
   you can paste from quickly.
2. Open **https://h2.nexus/cli** — click "free server for 15 minutes" and
   choose **Debian 11**.
3. Wait ~30 seconds for the VM to provision. Click into the **web console**.
4. Paste **ONE LINE** (replace `ghp_xxx` with your PAT):
   ```bash
   curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/h2-quick.sh | bash -s -- ghp_xxx
   ```
5. Wait ~12 minutes. Results push to `trikiman/vlessfilter` automatically.
   The VM auto-deletes at 15 min — no cleanup needed.

### Pros
- No new accounts, no card, no signup at all
- 4-CPU AMD EPYC + German peering = much faster than home connections
- Zero ongoing maintenance — just click + paste when you want fresh keys
- Russia-friendly host

### Cons
- Manual trigger required (no auto-schedule)
- Reduced run scope (5k untested batch vs 80k for full runs) to fit in
  15 minutes — full pool coverage requires multiple sessions
- Skips post-publish accuracy probe (saves time)

---

## Option D: Termux on Android phone
**Setup time:** ~10 minutes.
**Always-on:** Only when phone is plugged in + on home wifi.

### Steps

1. **Install Termux** from F-Droid: https://f-droid.org/en/packages/com.termux/
   (Play Store version is outdated — use F-Droid.)
2. **Open Termux** and run:
   ```bash
   pkg update && pkg upgrade -y
   pkg install -y golang git curl
   ```
3. **Generate a PAT** as in Option A step 3.
4. **Install vlessfilter**:
   ```bash
   curl -sSL https://raw.githubusercontent.com/trikiman/vlessfilter/main/scripts/install-always-on.sh \
     | bash -s -- github_pat_xxx
   ```
   Termux compatibility note: cron may fail to install — use
   `termux-job-scheduler` instead:
   ```bash
   pkg install -y termux-services
   termux-job-scheduler --period-ms 21600000 --script $HOME/.vlessfilter/refresh.sh
   ```
   (21600000 ms = 6 hours)
5. **Keep phone plugged in** and screen-off-but-not-sleeping for the
   pipeline to complete a 60-min run uninterrupted.

### Pros
- No new accounts, no card
- Uses hardware you already own

### Cons
- Phone must stay charging and not in deep sleep
- Heat from CPU during 60-min runs
- Battery wear over time

---

## Recommendation

| Your situation | Pick |
|----------------|------|
| Want zero-setup, just click+paste when you want fresh keys | **Option C** (h2.nexus) |
| Have a spare email + can use GitHub | **Option A** (GitHub Actions fork) |
| Russian-friendly Oracle / Google card works | **Option B** (Oracle Always-Free) |
| Don't want a new account or card check | **Option D** (Termux on Android) |

**Fastest path: try Option C (h2.nexus) RIGHT NOW** — no signup, just paste
a one-liner in their free VM. Refresh whenever you open the page. For
full automation later, do Option A or B.

## After deployment

The pipeline pushes results to `trikiman/vlessfilter` main branch. Your
subscription URLs stay the same:

```
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/all.txt
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/<CC>.txt
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vless/all.txt
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/vmess/all.txt
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/trojan/all.txt
https://raw.githubusercontent.com/trikiman/vlessfilter/main/subs/ss/all.txt
```

Local Windows scheduling is now disabled. You can remove it permanently:

```powershell
Unregister-ScheduledTask -TaskName "VlessFilter Refresh" -Confirm:$false
```
