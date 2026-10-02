# Student K3s Setup

This guide installs a single-node K3s server for the Archon Base enthusiast profile and starts a local resource monitor.

## 1. Check the host

The target practice profile is **8 GB RAM, 2 CPU cores, 1 GB graphics memory**, and at least 20 GB available under `/var/lib`. GPU memory is optional for CPU-only Kubernetes practice.

```bash
sudo ./scripts/install-k3s-student.sh --check-only
```

The check is read-only. It requires Linux, systemd, curl, and the standard host inspection tools.

## 2. Install K3s

Review the script before running it. The installer downloads the official K3s installer from `get.k3s.io`, installs a single-node server, enables the `k3s` systemd service, and keeps the admin kubeconfig at mode `0600`.

```bash
sudo ./scripts/install-k3s-student.sh
```

To inspect the command without changing the host:

```bash
sudo ./scripts/install-k3s-student.sh --dry-run
```

Verify the node:

```bash
sudo k3s kubectl get nodes
sudo systemctl status k3s
```

## 3. Build the monitor

```bash
go build -o ./bin/archon-monitor ./cmd/archon-monitor
sudo install -m 0755 ./bin/archon-monitor /usr/local/bin/archon-monitor
```

The monitor listens on `127.0.0.1:9105` and checks the K3s filesystem by default. Run it directly while learning:

```bash
sudo /usr/local/bin/archon-monitor \
  -listen 127.0.0.1:9105 \
  -data-dir /var/lib/archon-monitor \
  -disk-path /var/lib/rancher/k3s
```

The daemon exposes:

| Endpoint | Purpose |
|---|---|
| `/healthz` | Process health and whether a snapshot exists |
| `/status` | Latest JSON resource snapshot and profile compliance |
| `/metrics` | Prometheus-compatible local resource metrics |

A one-shot check is useful in scripts:

```bash
archon-monitor -once
```

The command exits with status `2` when the host is below the hard CPU, RAM, or disk recommendations. Missing NVIDIA GPU hardware is reported as a warning because CPU-only K3s practice remains supported.

## 4. Optional systemd service

The repository includes `deploy/systemd/archon-monitor.service`. Create an `archon` system user, install the binary, copy the unit, and enable it only after reviewing the hardening settings for your host:

```bash
sudo useradd --system --home /var/lib/archon-monitor --shell /usr/sbin/nologin archon
sudo install -D -m 0644 deploy/systemd/archon-monitor.service /etc/systemd/system/archon-monitor.service
sudo systemctl daemon-reload
sudo systemctl enable --now archon-monitor
```

Keep the monitor loopback-only unless an authenticated private-network proxy is configured. See [Cloud Security](../Cloud-Security/README.md) before using a cloud VM.
