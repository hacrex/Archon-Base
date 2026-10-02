# Self-hosting Requirements

Archon Base is designed to run on infrastructure you control, from a small local learning machine to a multi-node private AI cloud.

## Enthusiast / K3s practice tier

This is the recommended starting point for students and enthusiasts who want to practice self-hosting, Kubernetes concepts, K3s operations, and the Archon Base control plane on one machine.

| Resource | Minimum practice profile |
|---|---:|
| Memory | **8 GB RAM** |
| CPU | **2 CPU cores** |
| Graphics | **1 GB graphics memory** |
| Topology | Single node / local K3s |
| Intended use | Learning, development, lightweight workloads |

This tier is intended for the control plane, local platform experiments, and lightweight AI workloads. Model size, vector indexes, concurrent agents, and GPU-backed inference may require more resources. Add capacity as workloads grow rather than treating this profile as a production recommendation.

### Suggested student setup

- Linux host or Linux virtual machine
- K3s for a lightweight Kubernetes distribution
- Container runtime supported by K3s
- At least 20 GB of free disk space for images, local data, and logs
- Network access for downloading images and dependencies during setup

For a Kubernetes-free Phase 1 development path, use the [local execution specification](phase-1-local-execution-spec.md), which runs the API, PostgreSQL, and Qdrant through Docker Compose.

## Larger profiles

| Profile | CPU | RAM | Disk | Intended use |
|---|---:|---:|---:|---|
| Single node evaluation | 4 vCPU | 16 GB | 100 GB SSD | Evaluation and small deployments |
| Team cluster | 8 vCPU each | 32 GB each | 500 GB NVMe each | Multi-node development and team workloads |
| Production | Workload dependent | Vector memory typically 1.5–2× raw index size | NVMe for databases | Capacity-planned workloads |

These profiles describe starting points, not hard workload limits. Actual requirements depend on the selected model, vector engine, storage retention, number of agents, and concurrency.
