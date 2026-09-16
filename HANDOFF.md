# Handoff Document: VCert `pickupFirst` & Upstream PR #688

**Date**: 2026-09-04  
**Project**: [`vcert`](file:///C:/dev/vcert) (`github.com/venafi/vcert`)  
**Active Upstream PR**: [**https://github.com/Venafi/vcert/pull/688**](https://github.com/Venafi/vcert/pull/688)  
**Standalone Fork Repo**: [**https://github.com/tall27/vcert-pickupFirst**](https://github.com/tall27/vcert-pickupFirst)  
**PR Working Fork**: `tall27/vcert` (branch `add-pickup-first-mode`)

---

## 1. Executive Summary & Session Achievements

In this session, we completed the full implementation, validation, and upstream submission of **`pickupFirst` mode** ([GitHub Issue #649](https://github.com/Venafi/vcert/issues/649)):

1. **Full Multi-Platform Support**:
   - **Venafi TPP (Self-Hosted)**: O(1) metadata lookup via `RetrieveCertificateMetaData(dn)`.
   - **CyberArk Certificate Manager SaaS (Palo Alto Networks NGTS)**: Search by CN or pickupId, non-retired newest cert evaluation, token/client credentials.
   - **CyberArk Certificate Manager SaaS (Venafi Cloud / `vcp`)**: Search by CN or pickupId via API key, non-retired newest cert evaluation, local key binding for follower nodes.
2. **Cluster Convergence**:
   - Live tested on all three platforms: initial enrollment, cache/match skipping, downgrade refusal, and follower node convergence in under 1 second without duplicate certificate issuances.
3. **Upstream PR #688 Submitted**:
   - Target: `Venafi/vcert:master`
   - Head: `tall27/vcert:add-pickup-first-mode`
   - Scope: Exactly **14 core files** (+1,089 additions, -5 deletions), 0 merge conflicts.
4. **Documentation & Credentials**:
   - `CREDENTIAL_MANAGEMENT.md` updated with CyberArk Certificate Manager SaaS (API key) prominently at the top.
   - Example playbooks added for both `ngts` and `vcp`.
   - Standalone repo `tall27/vcert-pickupFirst` updated, tagged, and released with `v5.13.12-pickupFirst`.
   - Upstream PR branch `tall27/vcert:add-pickup-first-mode` updated with Contract 001 (central platform authority reconciliation).
   - Contract 001 formalized in `BACKLOG.md` ensuring full NIST SP 800-52/800-57 compliance.
   - Generated native `tldraw` flow diagram (`vcert-pickupfirst-flow.tldr`) and HTML preview (`vcert-pickupfirst-flow.html`).

---

## 2. GitHub Status on Upstream PR #688

When viewing [PR #688](https://github.com/Venafi/vcert/pull/688), GitHub displays:

```
Review required: At least 2 approving reviews are required by reviewers with write access.
Merging is blocked:
  - Commits must have verified signatures.
  - At least 2 approving reviews are required by reviewers with write access.
  - You're not authorized to push to this branch.
```

### Explanation of GitHub Checks:
1. **"At least 2 approving reviews are required by reviewers with write access"**:
   - This is normal branch protection on `Venafi/vcert:master`. Only upstream maintainers (Venafi/CyberArk engineers) can approve and merge.
2. **"You're not authorized to push to this branch"**:
   - This is normal. Contributors cannot push directly to `Venafi/vcert:master`. You push to your fork branch (`tall27:add-pickup-first-mode`), which automatically updates the PR.
3. **"Commits must have verified signatures"**:
   - The upstream repository `Venafi/vcert` enforces GPG/SSH signed commits. Commit `e17cdb7` was pushed without a cryptographic signature (`Verified` badge).

---

## 3. Action Items for the Next Session

### Step 1: Sign the Commit on Branch `add-pickup-first-mode`

To satisfy the `Commits must have verified signatures` rule:

#### Option A: Using SSH Signing Key (Fastest & Simplest)
1. Generate an SSH signing key (if not already existing):
   ```bash
   ssh-keygen -t ed25519 -C "tall27@users.noreply.github.com" -f ~/.ssh/id_ed25519_sign
   ```
2. Configure Git to use SSH for signing:
   ```bash
   git config --global gpg.format ssh
   git config --global user.signingkey ~/.ssh/id_ed25519_sign.pub
   git config --global commit.gpgsign true
   ```
3. Add the public key (`~/.ssh/id_ed25519_sign.pub`) to GitHub:
   - Go to: **GitHub Settings -> SSH and GPG keys -> New SSH key**
   - Key type: Select **Signing Key** (NOT Authentication Key).
4. Re-sign the commit in `C:\dev\vcert-pr`:
   ```bash
   cd C:\dev\vcert-pr
   git commit --amend -S --no-edit
   git push origin add-pickup-first-mode --force
   ```
   GitHub will immediately mark the commit as **`Verified`** and clear the signature block!

#### Option B: Using GPG Key
1. Install Gpg4win / GPG on Windows.
2. Generate a GPG key for `tall27@users.noreply.github.com`.
3. Add the public GPG key to GitHub Settings -> SSH and GPG keys.
4. Run `git commit --amend -S --no-edit` and force push to `origin add-pickup-first-mode`.

---

### Step 2: Request Review from Maintainers
Once the commit is signed:
1. In PR #688 comments, leave a brief note for maintainers (e.g. pinging Jeremy Meldrum or repo maintainers), referencing Issue #649 and PR #650:
   ```markdown
   Hi @jmeldrum76 and @Venafi maintainers - This PR completes the implementation of #649 originally prototyped in #650, extending native `pickupFirst` support across TPP, NGTS, and CyberArk Certificate Manager SaaS with full unit tests and live verification. Looking forward to your review!
   ```

---

## 4. Key Directory & Repository Locations

| Directory / Resource | Description |
|---|---|
| `C:\dev\vcert` | Local working repository (`master` tracking upstream `venafi/vcert`). |
| `C:\dev\vcert-pr` | Dedicated clean clone of `tall27/vcert` on branch `add-pickup-first-mode` (backing PR #688). |
| `C:\dev\vcert-github-repo` | Local clone of `tall27/vcert-pickupFirst` (standalone audited release repo). |
| `C:\dev\vcert\vcert-pickupFirst-dist.zip` | Standalone distribution archive (329 files, clean of git/binaries). |
| [**Venafi/vcert #688**](https://github.com/Venafi/vcert/pull/688) | Upstream Pull Request on official repository. |
| [**tall27/vcert-pickupFirst**](https://github.com/tall27/vcert-pickupFirst) | Public standalone repository with releases. |

---

## 5. Summary of Files in PR #688 (Exact 14 Files)

```
README-PLAYBOOK.md                           (+83/-0)
examples/playbook/ngts_pickup_first.yaml     (+24/-0)
examples/playbook/vcp_pickup_first.yaml      (+19/-0)
pkg/playbook/app/domain/playbookRequest.go   (+2/-0)
pkg/playbook/app/installer/crypto.go         (+3/-0)
pkg/playbook/app/service/pickup_first.go     (+172/-0)
pkg/playbook/app/service/pickup_first_test.go(+139/-0)
pkg/playbook/app/service/service.go          (+8/-0)
pkg/playbook/app/vcertutil/vcertutil.go      (+367/-0)
pkg/playbook/app/vcertutil/vcertutil_test.go (+152/-0)
pkg/venafi/cloud/connector.go                (+27/-1)
pkg/venafi/cloud/search.go                   (+1/-0)
pkg/venafi/ngts/connector.go                 (+91/-4)
pkg/venafi/ngts/search.go                    (+1/-0)
```
