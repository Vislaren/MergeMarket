# VPS Setup

> Captures the manual VPS provisioning completed before the agent sessions
> began (task A-03). Recorded here for reference per the DONE.md note.

## Host

- **Provider:** Contabo VPS
- **OS:** Ubuntu 22.04 LTS
- **Public IP:** `95.111.228.35`

## Installed Stack

| Component | Purpose |
|---|---|
| Docker | Container runtime |
| K3s | Lightweight single-node Kubernetes orchestration |
| Helm | Kubernetes package manager |
| Terraform | Infrastructure-as-Code (`infra/terraform/`) |

K3s is running and the cluster node reports `Ready`.

## Services Reachable on the VPS

| Service | URL | Set up in |
|---|---|---|
| SonarQube Community | http://95.111.228.35:9000 | A-11 |
| Grafana | http://95.111.228.35:3000 | A-12 |

## Key Decisions

- **K3s over full Kubernetes** — single-node operation on a modest VPS;
  lower memory footprint, simpler to manage.
- **Docker** as the container runtime for both local dev and the VPS.

## Known Limitations / Follow-ups

- This setup was performed manually outside the agent workflow; no
  Terraform state yet captures it. Codifying the VPS baseline in
  `infra/terraform/` is a future task.
- SonarQube (A-11) and Grafana (A-12) are not yet deployed; the URLs above
  are the intended endpoints.

## Firewall Ports (intended)

| Port | Service |
|---|---|
| 22 | SSH |
| 80 / 443 | NGINX (CrowdSec WAF in front) |
| 9000 | SonarQube (opened in A-11) |
| 3000 | Grafana (opened in A-12) |
