# Agent A — Done

> This file is the permanent record of everything Agent A has completed.
> Add one entry per completed task. Never delete entries.
> Agent B reads this file to know what to test each session.

---

## Completed Tasks

_No tasks completed yet. The project has not started._

---

### [DONE] A-03 — VPS Setup
**Session:** Pre-project (completed manually before agent sessions began)
**Completed by:** User (manual setup)
**Commit:** N/A — done outside agent workflow

**What was built:**
VPS fully set up with Docker, K3s, Helm, and Terraform installed on the
Ubuntu 22.04 Contabo VPS. K3s is running and the cluster node is Ready.

**Files created/modified:**
- Manual setup — no files committed yet. Agent A should document this
  in `docs/vps-setup.md` during session A-01 or A-02 for reference.

**Key decisions made:**
- Using Contabo VPS with Ubuntu 22.04
- K3s chosen over full Kubernetes for single-node operation
- Docker installed for container runtime

**API contracts changed:** No

**Known limitations:**
- Setup not yet documented in the repo. Should be captured in
  `docs/vps-setup.md` during an early session.

---

### [DONE] A-01 — Initialise Repo Structure
**Session:** 1
**Completed by:** Agent A
**Commit:** `session(A-01): initialise repo structure` _(local commit — push deferred, see limitations)_

**What was built:**
The full MergeMarket folder layout as defined in
`project_docs/architecture/ARCHITECTURE.md §5` (the doc referenced in
INSTRUCTIONS as `Documentation.md §6.2` does not exist — §5 of ARCHITECTURE
is the authoritative structure). Added a root `.gitignore` covering Go,
Flutter, Terraform, SonarQube, and environment files. Added descriptive
`README.md` files in each service folder and `.gitkeep` placeholders in
every otherwise-empty directory so Git tracks the full tree. Also captured
the manual A-03 VPS setup in `docs/vps-setup.md` (per the DONE.md note on
the A-03 entry) and added a root project `README.md`.

**Files created/modified:**
- `.gitignore`
- `README.md`
- `docs/vps-setup.md`
- `api-gateway/README.md`, `api-gateway/plugins/.gitkeep`
- `services/{auth,bff,scraper-service,normalization,history,proxy-validator,mock-server}/README.md`
- `services/scraper-service/configs/.gitkeep`
- `apps/mobile/lib/{screens,widgets,providers,services,models,theme}/.gitkeep`
- `apps/mobile/test/{unit,mocks}/.gitkeep`, `apps/mobile/integration_test/.gitkeep`
- `infra/{k3s,terraform}/.gitkeep`, `shared/.gitkeep`, `docs/testing/.gitkeep`

**Key decisions made:**
- Used `.gitkeep` for empty leaf directories (Git does not track empty
  folders); service folders got real one-paragraph READMEs instead of empty
  files, which is more useful while still satisfying trackability.
- `.gitignore` whitelists `.env.example` and `*.tfvars.example` so the
  example files (A-02) remain committable while real secrets are ignored.
- Treated `project_docs/architecture/ARCHITECTURE.md §5` as the source of
  truth for the folder layout since `Documentation.md` is absent.

**API contracts changed:** No

**Known limitations / follow-ups for the user:**
- **Push deferred — repo not pushable yet.** Two blockers: (1) no Git
  remote is configured and no repository URL was provided (INSTRUCTIONS §7
  expected one at the first session); (2) the Git root is `C:/project`, a
  shared monorepo also containing unrelated projects (SafeZone, E_commerce,
  etc.), **not** a dedicated MergeMarket repo. Decide whether to (a) create
  a dedicated repo rooted at `C:/project/mergemarket`, or (b) push the
  shared repo to a provided remote. Then the §7 push protocol can run.
- A draft `.github/workflows/ci.yml` already exists at
  `mergemarket/.github/workflows/`. GitHub Actions only reads workflows at
  the **repo root**, so from `C:/project` this draft will not trigger. This
  is A-10's concern; flagged here so it isn't lost.

---

## Entry Format

When you complete a task, add an entry using this format:

```markdown
---

### [DONE] A-XX — Task Name
**Session:** X
**Completed by:** Agent A
**Commit:** `session(A-XX): task name`

**What was built:**
Brief description of what was created or configured.

**Files created/modified:**
- `path/to/file1`
- `path/to/file2`

**Key decisions made:**
Any architectural or implementation decisions and the reason for them.

**API contracts changed:**
Yes / No — if Yes, describe what changed so Agent B can update mocks.

**Known limitations:**
Anything left incomplete, deferred, or that needs follow-up.
```
