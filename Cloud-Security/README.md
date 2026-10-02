# Cloud VM Security Baseline

The student K3s profile is suitable for a lab VM or a personal machine, but it should not be exposed directly to the public internet.

## Before installing K3s

- Use a supported Linux distribution with security updates enabled.
- Create a non-root operator account with SSH keys; use `sudo` only for installation.
- Keep SSH restricted to a trusted source IP, VPN, or bastion host.
- Enable a host firewall and allow only SSH plus the specific local development ports required by the lab.
- Do not expose the K3s API, PostgreSQL, Qdrant, or the monitor port (`9105`) publicly.
- Keep backups of important manifests and data outside the VM.

## K3s and kubeconfig

- Keep `/etc/rancher/k3s/k3s.yaml` mode `0600`.
- Treat kubeconfig as a cluster-admin credential; never commit it, paste it into issues, or place it in a public artifact store.
- Use `k3s kubectl` or a copied user kubeconfig through a protected local path.
- Review installed workloads and remove lab resources that are no longer needed.

## Resource monitor

- The monitor binds to `127.0.0.1:9105` by default and writes its snapshot with mode `0600`.
- If remote dashboards are needed, put an authenticated reverse proxy or private network in front of `/metrics`; do not bind the daemon to `0.0.0.0` by default.
- The monitor reads local resource metadata only. Keep logs and snapshots on a private filesystem.

## Cloud provider notes

Apply the provider's equivalent network controls and instance hardening for AWS security groups, Azure NSGs, Google Cloud VPC firewall rules, or another provider's firewall layer. Cloud firewall rules do not replace host firewall rules; use both layers.
