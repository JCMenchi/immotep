# Immotep — Installation Guide

This document summarizes available installation methods and the minimal commands to deploy Immotep from this repository.

## Overview

Available options (dir under `install/`):

- `systemd/` — local installer: shell script + systemd unit.
- `k8s/` — Kubernetes manifests and helper scripts,if you have access to K8S cluster adn docker registry.
- `ansible/` — repeatable provisioning with Ansible playbook and templates for ubuntu OS.

Cloud helpers (to use with ansible):

- `gcp/` — Google Cloud helper scripts.
- `ovh/` — OVH Cloud provider scripts.

## Prerequisites

- Linux server (Ubuntu).
- For k8s: docker, kubectl and access to a cluster.

## Install

### 1. systemd (quick local install)

Path: `install/systemd/`

Main script: [install.sh](./systemd/install.sh) — performs user/group creation, copies files, sets capabilities, creates DB and extensions, enables the systemd service.

Run (from repo root):

```bash
cd install/systemd
sudo install.sh
```

### 2. Kubernetes (k8s)

Path: `install/k8s/`

*Description:* Use this method for deploying Immotep in a Kubernetes cluster. It includes manifests and helper scripts for setting up the application. K8S manifest shall be adapted to your context.

*Steps:*

1. Ensure you have Docker, kubectl, and access to a Kubernetes cluster,
2. Navigate to the k8s directory,
3. Deploy PostgreSQL databse: use [create_services.sh](./k8s/create_services.sh) to build an OCI image with postgres; and install in the K8S cluster,
4. Adapt and Apply the Kubernetes manifests: [k8s_deploy.yaml](./k8s/k8s_deploy.yaml).

```bash
cd install/k8s
./create_services.sh
kubectl apply -f k8s_deploy.yaml
```

### 3. Ansible

Path: `install/ansible/`

*Description:* This method is designed for repeatable provisioning using Ansible playbooks and templates. It is suitable for managing multiple servers.

*Steps:*

1. Create a server in some cloud provider, (see examples for [GCP](./gcp/README.md) and [OVH](./gcp/README.md)),
2. Navigate to the ansible directory,
3. Run the Ansible playbook,

```bash
cd install/ansible
ansible-playbook -i inventory.yml immotep.yml
```
