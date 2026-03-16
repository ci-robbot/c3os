# Kairos SLE Micro 5.5

This example demonstrates how to build a Kairos image based on SUSE Linux Enterprise Micro (SLEM) 5.5.

## Prerequisites

- Docker or Podman
- Access to SUSE repositories (you need to provide the repo files)

## Building

1. Prepare your repository files:
   - Create a `repos/` directory with your SUSE and openSUSE repository files
   - Create a `services/` directory with your zypper service files

2. Build the image:
   ```bash
   docker build -t kairos-slemicro-5.5 .
   ```

## Repository Requirements

The Dockerfile expects the following files to be present:

- `repos/SUSE*` - SUSE package repository files
- `repos/opensuse*` - openSUSE package repository files  
- `services/*` - Zypper service definition files

## What this example does

- Installs essential packages like cryptsetup, nfs-client, systemd-network, etc.
- Adds monitoring and security tools (htop, fail2ban, rsyslog, etc.)
- Removes elemental packages and cleans up related services
- Generates initrd for booting

## Notes

- You need valid SUSE subscription/repo access to build this image
- The example uses `registry.suse.com/suse/sle-micro/5.5:latest` as the base image
