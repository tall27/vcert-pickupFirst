# VCert — Multi-Node Shared Certificate Convergence (`pickupFirst`)

> **Automated, idempotent certificate and private key distribution across server clusters for Venafi TPP & CyberArk Certificate Manager SaaS (NGTS).**

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Issue](https://img.shields.io/badge/GitHub%20Issue-%23649-orange.svg)](https://github.com/Venafi/vcert/issues/649)
[![Platforms](https://img.shields.io/badge/Platforms-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](#building-from-source)
[![Upstream](https://img.shields.io/badge/Upstream-Venafi%2Fvcert-blue)](https://github.com/Venafi/vcert)

---

### Reference to Upstream VCert

This repository is an enhanced distribution of the official [**Venafi VCert** (`Venafi/vcert`)](https://github.com/Venafi/vcert).

It implements the feature requested in [**GitHub Issue #649**](https://github.com/Venafi/vcert/issues/649): the **`pickupFirst`** playbook engine, with end-to-end support for both **Venafi Trust Protection Platform (TPP)** and **CyberArk Certificate Manager SaaS / Venafi Next-Gen Trust Security (NGTS)**.

* **Upstream Repository**: [github.com/Venafi/vcert](https://github.com/Venafi/vcert)
* **Official Documentation**: [Venafi Documentation](https://docs.venafi.com) | [CyberArk Certificate Manager Documentation](https://docs.cyberark.com)
* **Status**: Fully backward-compatible drop-in replacement with enhanced multi-node playbook coordination.

---

## The Problem: Clustered Certificate Distribution

In clustered or load-balanced topologies (e.g. Node 1 and Node 2 serving the same hostname or wildcard), multiple servers require the exact same certificate and private key:

* **Traditional Playbook Behavior**: Every node independently executes certificate enrollment. This creates duplicate certificates, wastes CA quotas and budget, and leaves nodes with mismatched private keys.
* **Complex Workarounds**: Teams historically built bespoke rsync scripts, shared NFS mounts, SSH copy routines, or maintained separate "Leader" vs "Follower" playbooks.

---

## The Solution: `pickupFirst: true`

With `pickupFirst: true`, **all nodes run the exact same command with the exact same playbook**:

```bash
vcert run -f playbook.yaml
```

No custom wrapper scripts. No separate leader/follower configurations. Zero coordination infrastructure required.

### Identical Playbook Configuration

```yaml
certificateTasks:
  - name: shared-service-cert
    renewBefore: 1d
    request:
      csr: service
      pickupFirst: true            # <-- Enables automatic multi-node convergence
      zone: 'Private 5 days'
      subject:
        commonName: 'shared.example.com'
    installations:
      - format: PEM
        file: /etc/ssl/certs/app.crt
        keyFile: /etc/ssl/private/app.key
        chainFile: /etc/ssl/certs/chain.crt
        afterInstallAction: "systemctl reload nginx"
```

---

## How the Nodes Coordinate Automatically

```
                   vcert run -f playbook.yaml
                              │
                 Does cert exist on Platform?
                             ╱ ╲
                       YES ╱     ╲ NO
                         ╱         ╲
      Is Platform newer             Node 1 (First to run):
     than installed cert?           - Falls through & Enrolls cert
            ╱ ╲                     - Key generated on Platform (csr: service)
      YES ╱     ╲ NO (Match)        - Installs cert + key locally
        ╱         ╲
 Node 2 (Follower):  Both Nodes (Cron / Next run):
 - Downloads cert    - Thumbprints match
   + key from DEK    - Checks renewBefore window
 - Installs locally  - If healthy: Exits in < 1s with NO action
 - ZERO enrollment
```

### Execution Lifecycle

1. **Whichever node runs first (e.g. Node 1 - The Pioneer)**:
   * Queries the platform for `shared.example.com`.
   * Finds nothing &rarr; automatically falls through to enroll it.
   * The platform generates the private key (`csr: service`) and issues the certificate.
   * Node 1 installs the certificate, private key, and chain locally, then executes `afterInstallAction`.

2. **The other node (Node 2 - The Follower)**:
   * Runs the **identical command**: `vcert run -f playbook.yaml`.
   * Queries the platform for `shared.example.com`.
   * Finds the certificate already enrolled by Node 1.
   * Since Node 2 has no certificate installed yet, it downloads the certificate and the vaulted private key (via secure DEK decryption) and installs them.
   * **Zero duplicate enrollment requests are created.**

3. **Subsequent runs on all nodes (e.g. automated daily cron)**:
   * Both nodes run `vcert run -f playbook.yaml`.
   * Both inspect their local certificate: the SHA-1 thumbprint matches the platform certificate.
   * Both evaluate the `renewBefore` window. If the certificate is healthy, both exit in **< 1 second** with *"certificate in good health. No actions needed"*.
   * When the renewal window eventually arrives, whichever node executes first renews the certificate. The remaining nodes automatically pick up the new certificate and private key on their next run.

---

## Decision Engine Matrix

Before triggering any certificate request, VCert evaluates a 4-way decision matrix:

| State | Condition | VCert Action |
|---|---|---|
| **Match** | Local thumbprint == Platform thumbprint | Defer to `renewBefore` window check. Exit in <1s if healthy. |
| **Platform Newer** | Platform NotAfter > Local NotAfter (or no local cert) | Download certificate + private key + chain, install, trigger `afterInstallAction`. Skip enrollment. |
| **Platform Older** | Platform NotAfter < Local NotAfter | Refuse downgrade with a warning log. Exit cleanly. |
| **Not Found** | No certificate matching Common Name or `pickupId` | Fall through to standard certificate enrollment. |

---

## Quickstart & Verification

### Using Pre-built Windows Binary
A pre-compiled 64-bit Windows binary is included in the release:

```powershell
.\vcert.exe --version
# Output: vcert.exe version v5.13.9-pickupFirst

.\vcert.exe run -f ./playbook.yaml
```

### Building from Source

You can compile VCert for any operating platform (Linux, macOS, Windows) using Go 1.21+:

#### Linux (x86_64 / amd64)
```bash
GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
chmod +x vcert
./vcert --version
```

#### Linux (ARM64)
```bash
GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
```

#### macOS (Apple Silicon / arm64)
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
```

#### macOS (Intel / amd64)
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
```

#### Windows (amd64)
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert.exe ./cmd/vcert
```

### Running Unit Tests
```bash
go test -v ./pkg/playbook/app/service -run TestPickupFirst
go test -v ./pkg/playbook/app/vcertutil -run Test
```

---

## Platform Support

| Platform | Supported | Private Key Retrieval | Discovery Method |
|---|:---:|---|---|
| **Venafi TPP** (Self-Hosted) | Yes | Vaulted key via TPP WebSDK (`/vedsdk/Certificates/Retrieve`) | DN search: `<zone>\<commonName>` or custom `pickupId` |
| **CyberArk Certificate Manager SaaS (NGTS)** | Yes | Data Encryption Key (DEK) decryption via `/v1/certificates/{id}/privatekey` | CN search: `/v1/certificates?filter=subjectCN:eq:...` or fingerprint |

---

## Documentation & Guides

* **Interactive Guide**: Open [`pickup_first_guide.html`](pickup_first_guide.html) in your browser for a visual topology walkthrough.
* **Playbook Reference**: See [`README-PLAYBOOK.md`](README-PLAYBOOK.md) for full configuration specifications.
* **Distribution Guide**: See [`README-DISTRIBUTION.md`](README-DISTRIBUTION.md) for architecture details.
* **Platform CLIs**:
  * [CyberArk Certificate Manager, SaaS (NGTS)](README-CLI-NGTS.md)
  * [CyberArk Certificate Manager, Self-Hosted (TPP)](README-CLI-PLATFORM.md)
  * [CyberArk Workload Identity Manager (Firefly)](README-CLI-FIREFLY.md)

---

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
