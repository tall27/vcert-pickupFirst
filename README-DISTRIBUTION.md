# VCert with pickupFirst (GitHub Issue #649) & Full NGTS Support

This distribution package contains the updated **VCert CLI** with support for the **`pickupFirst`** playbook feature across both **Venafi TPP** and **Venafi NGTS (CyberArk Certificate Manager SaaS)**.

Included in this package:
1. **Pre-built Windows x64 binary**: `vcert.exe` (Version: `v5.13.9-pickupFirst`)
2. **Complete Go source code**: Ready for audit and multi-platform compilation (`linux/amd64`, `linux/arm64`, `darwin/arm64`, etc.)
3. **Interactive Guide**: `pickup_first_guide.html` (Open in any modern browser)
4. **Configuration Examples**: `examples/playbook/ngts_pickup_first.yaml` and `README-PLAYBOOK.md`

---

## 1. Quick Verification (Pre-built Windows Binary)

To check the pre-built binary on Windows:

```powershell
.\vcert.exe --version
# Expected output: vcert.exe version v5.13.9-pickupFirst
```

Display playbook run help:

```powershell
.\vcert.exe run --help
```

---

## 2. Building from Source for Any Operating Platform

The full Go source code is included so you can compile VCert for your specific target operating system and architecture.

### Prerequisites
- **Go**: Version 1.21 or higher installed (`go version`)

### Build Commands

#### Linux (x86_64 / amd64)
```bash
GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
chmod +x vcert
./vcert --version
```

#### Linux (ARM64)
```bash
GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
chmod +x vcert
./vcert --version
```

#### macOS (Apple Silicon / M1 / M2 / M3)
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
chmod +x vcert
./vcert --version
```

#### macOS (Intel / amd64)
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert ./cmd/vcert
chmod +x vcert
./vcert --version
```

#### Windows (x86_64)
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -ldflags "-X github.com/Venafi/vcert/v5.versionString=v5.13.9-pickupFirst -s -w" -o vcert.exe ./cmd/vcert
.\vcert.exe --version
```

---

## 3. What is `pickupFirst`? (Issue #649)

### Why It Works with the Same Command

That is the entire purpose of `pickupFirst: true`. You do not need special flags, separate configurations, or custom wrapper scripts for different nodes.

The playbook file is identical on every machine:

```yaml
certificateTasks:
  - name: shared-service-cert
    renewBefore: 1d
    request:
      csr: service
      pickupFirst: true            # <-- Enables automatic convergence
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

### How the Nodes Coordinate Automatically

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

1. **Whichever node runs first (e.g. Node 1)**:
   - Queries the platform for `shared.example.com`.
   - Finds nothing &rarr; automatically falls through to enroll it.
   - The platform generates the private key (`csr: service`) and issues the certificate.
   - Node 1 installs the certificate and private key.

2. **The other node (Node 2)**:
   - Runs the identical command: `vcert run -f playbook.yaml`.
   - Queries the platform for `shared.example.com`.
   - Finds the certificate Node 1 enrolled.
   - Since Node 2 has no certificate installed yet, it downloads the certificate and the server-side private key (via DEK decryption) and installs them.
   - Zero enrollment requests are created.

3. **Subsequent runs on both nodes (e.g. daily cron)**:
   - Both nodes run `vcert run -f playbook.yaml`.
   - Both inspect their local certificate: the thumbprint matches the platform.
   - Both check `renewBefore`: if still valid, both exit in <1 second with *"certificate in good health. No actions needed"*.
   - When the renewal window eventually hits, whichever node runs first renews it; the other node picks up the renewed certificate on its next run.

---

## 4. Playbook Configuration Syntax

Add `pickupFirst: true` under the `certificate:` block of any playbook.

### Venafi NGTS (CyberArk Certificate Manager SaaS) Example
```yaml
config:
  connection:
    platform: ngts
    url: "https://vaas.cyberark.cloud" # Or your tenant URL
    apiKey: "${VCERT_NGTS_API_KEY}"
    zone: "Default" # Or application/issuing template name

certificates:
  - name: shared-cluster-cert
    pickupFirst: true             # <--- Enables smart pickup-first workflow
    # pickupId: "custom-cn.example.com" # Optional override (defaults to commonName)
    commonName: "api-cluster.example.com"
    sanDNS:
      - "api-cluster.example.com"
      - "api-node1.example.com"
      - "api-node2.example.com"
    keyType: rsa
    keySize: 2048
    installations:
      - format: pem
        file: "/etc/ssl/certs/api-cluster.crt"
        keyFile: "/etc/ssl/private/api-cluster.key"
        afterInstall:
          - "systemctl reload nginx"
```

### Venafi TPP Example
```yaml
config:
  connection:
    platform: tpp
    url: "https://tpp.example.com/vedsdk"
    accessToken: "${VCERT_TPP_TOKEN}"
    zone: "DevOps\\Certificates"

certificates:
  - name: shared-cluster-cert
    pickupFirst: true
    commonName: "api-cluster.example.com"
    installations:
      - format: pem
        file: "/etc/ssl/certs/api-cluster.crt"
        keyFile: "/etc/ssl/private/api-cluster.key"
```

---

## 5. Code Changes & Architecture Audit

The following table outlines the key source code files modified or added for this feature:

| File | Type | Description |
|---|---|---|
| `pkg/playbook/app/domain/playbookRequest.go` | Modified | Added `PickupFirst bool` and `PickupID string` fields to `PlaybookRequestCertificate`. |
| `pkg/playbook/app/service/pickup_first.go` | **New** | Core 4-way decision matrix (`Match`, `PlatformNewer`, `PlatformOlder`, `NotFound`). |
| `pkg/playbook/app/service/service.go` | Modified | Integrated `pickupFirstAttempt` before enrollment check in `Execute()`. |
| `pkg/playbook/app/vcertutil/vcertutil.go` | Modified | Added `LocateResult`, multi-platform `LocateLatestCN`, `PickupCertificateByLocator`, and `NormalizeThumbprint`. |
| `pkg/venafi/ngts/connector.go` | Modified | Added `SearchCertificatesByCN`, `SearchCertificatesByFingerprint`, and `GetCertificateDetails`. |
| `pkg/venafi/ngts/search.go` | Modified | Added `CertificateStatus` field to NGTS `Certificate` struct. |
| `pkg/playbook/app/installer/crypto.go` | Modified | Added `LoadInstalledPEM` for inspecting locally installed certificate expiration & SHA-1 thumbprint. |
| `pkg/playbook/app/service/pickup_first_test.go` | **New** | Unit test suite for `pickupFirst` states and flags. |
| `pkg/playbook/app/vcertutil/vcertutil_test.go` | **New** | Unit test suite for locator helpers. |
| `pickup_first_guide.html` | **New** | Standalone interactive documentation guide (Light theme). |
| `README-PLAYBOOK.md` | Modified | Updated playbook specification documentation for `pickupFirst`. |

---

## 6. Running Tests

To run the unit tests included in this distribution:

```bash
go test -v ./pkg/playbook/app/service -run TestPickupFirst
go test -v ./pkg/playbook/app/vcertutil -run Test
```

All tests execute offline and verify decision matrices, edge cases, and helper utilities.
