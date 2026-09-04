# Credential Configuration & Run Guide

This guide explains from the very beginning how to configure, store, and use credentials to run **VCert** with:
1. **CyberArk Certificate Manager, SaaS (Venafi Cloud / `vcp`)** using **API Key** authentication *(Top Section)*
2. **CyberArk Certificate Manager, SaaS (Palo Alto Networks NGTS / `ngts`)** using **OAuth 2.0** authentication

---

## 1. CyberArk Certificate Manager SaaS (API Key Authentication)

The SaaS platform (`platform: vcp` / `cloud`) authenticates directly using an API key.

### Required Parameters

| Parameter | Flag | Environment Variable | Description / Example |
|---|---|---|---|
| **Platform** | `-p, --platform` | `VCERT_PLATFORM` | `vcp` (or `cloud`) |
| **API Key** | `-k, --apiKey` | `VCERT_APIKEY` | `<YOUR_SAAS_API_KEY>` |
| **Zone (Application \ Template)** | `-z, --zone` | `VCERT_ZONE` | `<APPLICATION_NAME>\<ISSUING_TEMPLATE>` |
| **Base URL** *(Optional)* | `-u, --url` | `VCERT_URL` | `https://api.venafi.cloud` *(default)* |

---

### Configuration Methods (SaaS / API Key)

#### Method A: Environment Variables

Set environment variables once so VCert CLI and playbooks authenticate automatically:

**Windows (PowerShell):**
```powershell
$env:VCERT_PLATFORM = "vcp"
$env:VCERT_APIKEY   = "<YOUR_SAAS_API_KEY>"
$env:VCERT_ZONE     = "<APPLICATION_NAME>\<ISSUING_TEMPLATE>"
```

**Linux / macOS (Bash):**
```bash
export VCERT_PLATFORM="vcp"
export VCERT_APIKEY="<YOUR_SAAS_API_KEY>"
export VCERT_ZONE="<APPLICATION_NAME>\\<ISSUING_TEMPLATE>"
```

#### Method B: Playbook YAML Configuration (`pickupFirst`)

For automated multi-node certificate deployment with `pickupFirst: true`, store connection parameters in the `config.connection` block:

```yaml
config:
  connection:
    platform: vcp
    credentials:
      apiKey: "${VCERT_APIKEY}"

certificateTasks:
  - name: shared-service-cert
    renewBefore: 5d
    request:
      pickupFirst: true                 # Automatic cluster convergence (GitHub Issue #649)
      zone: '<APPLICATION_NAME>\<ISSUING_TEMPLATE>'
      subject:
        commonName: 'shared.example.com'
    installations:
      - format: PEM
        file: /etc/ssl/certs/app.crt
        keyFile: /etc/ssl/private/app.key
        chainFile: /etc/ssl/certs/chain.crt
        afterInstallAction: "systemctl reload nginx"
```

To run:
```bash
vcert run -f ./playbook.yaml
```

#### Method C: Direct CLI Commands

**Enroll a Certificate:**
```bash
vcert enroll -p vcp -k "<YOUR_SAAS_API_KEY>" \
  -z "<APPLICATION_NAME>\<ISSUING_TEMPLATE>" \
  --cn "web.example.com" \
  --cert-file "./cert.pem" \
  --key-file "./cert.key" \
  --chain-file "./chain.pem" \
  --no-prompt
```

**Pickup an Existing Certificate:**
```bash
vcert pickup -p vcp -k "<YOUR_SAAS_API_KEY>" \
  --pickup-id "<CERTIFICATE_UUID_OR_FINGERPRINT>" \
  --cert-file "./cert.pem"
```

---

## 2. CyberArk Certificate Manager SaaS / NGTS (OAuth 2.0 Authentication)

The NGTS platform (`platform: ngts`) uses OAuth 2.0 Client Credentials.

### Required NGTS Parameters

| Parameter | Flag | Environment Variable | Example |
|---|---|---|---|
| **API Base URL** | `-u, --url` | `VCERT_URL` | `https://api.strata.paloaltonetworks.com/ngts` |
| **Token URL** | `--token-url` | `VCERT_TOKEN_URL` | `https://auth.apps.paloaltonetworks.com/am/oauth2/access_token` |
| **Client ID** | `--client-id` | `VCERT_CLIENT_ID` | `your-service-account-client-id` |
| **Client Secret** | `--client-secret` | `VCERT_CLIENT_SECRET` | `your-service-account-client-secret` |
| **Scope (TSG ID)** | `--scope` | `VCERT_SCOPE` | `tsg_id:<10_DIGIT_TSG_ID>` |
| **Issuing Template (Zone)** | `-z, --zone` | `VCERT_ZONE` | `<YOUR_ISSUING_TEMPLATE_NAME>` |

> [!NOTE]
> The `--scope` value must always follow the format `tsg_id:<10-digit-id>`, where `<10-digit-id>` is your Tenant Service Group identifier.

---

### Configuration Methods (NGTS / OAuth)

#### Method A: Environment Variables

**Windows (PowerShell):**
```powershell
$env:VCERT_PLATFORM      = "ngts"
$env:VCERT_URL           = "https://api.strata.paloaltonetworks.com/ngts"
$env:VCERT_TOKEN_URL     = "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token"
$env:VCERT_CLIENT_ID     = "your-service-account-client-id"
$env:VCERT_CLIENT_SECRET = "your-service-account-client-secret"
$env:VCERT_SCOPE         = "tsg_id:<10_DIGIT_TSG_ID>"
$env:VCERT_ZONE          = "<YOUR_ISSUING_TEMPLATE_NAME>"
```

**Linux / macOS (Bash):**
```bash
export VCERT_PLATFORM="ngts"
export VCERT_URL="https://api.strata.paloaltonetworks.com/ngts"
export VCERT_TOKEN_URL="https://auth.apps.paloaltonetworks.com/am/oauth2/access_token"
export VCERT_CLIENT_ID="your-service-account-client-id"
export VCERT_CLIENT_SECRET="your-service-account-client-secret"
export VCERT_SCOPE="tsg_id:<10_DIGIT_TSG_ID>"
export VCERT_ZONE="<YOUR_ISSUING_TEMPLATE_NAME>"
```

#### Method B: Configuration File (`vcert.ini`)

```ini
[default]
platform = ngts
url = https://api.strata.paloaltonetworks.com/ngts
ngts_token_url = https://auth.apps.paloaltonetworks.com/am/oauth2/access_token
ngts_client_id = your-service-account-client-id
ngts_client_secret = your-service-account-client-secret
ngts_scope = tsg_id:<10_DIGIT_TSG_ID>
```

Run commands with the INI file:
```bash
vcert enroll --config ./vcert.ini -z "<YOUR_ISSUING_TEMPLATE_NAME>" --cn app.example.com --no-prompt
```

#### Method C: Playbook YAML Configuration (`pickupFirst`)

```yaml
config:
  connection:
    platform: ngts
    url: "https://api.strata.paloaltonetworks.com/ngts"
    tokenUrl: "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token"
    clientId: "${VCERT_CLIENT_ID}"
    clientSecret: "${VCERT_CLIENT_SECRET}"
    scope: "tsg_id:<10_DIGIT_TSG_ID>"

certificateTasks:
  - name: shared-service-cert
    renewBefore: 1d
    request:
      csr: service
      pickupFirst: true                 # Automatic cluster convergence (GitHub Issue #649)
      zone: '<YOUR_ISSUING_TEMPLATE_NAME>'
      subject:
        commonName: 'shared.example.com'
    installations:
      - format: PEM
        file: /etc/ssl/certs/app.crt
        keyFile: /etc/ssl/private/app.key
        chainFile: /etc/ssl/certs/chain.crt
        afterInstallAction: "systemctl reload nginx"
```

#### Method D: OAuth Access Token Caching with `getcred`

Generate a short-lived access token:
```bash
vcert getcred -p ngts \
  --token-url "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token" \
  --client-id "$VCERT_CLIENT_ID" \
  --client-secret "$VCERT_CLIENT_SECRET" \
  --scope "tsg_id:<10_DIGIT_TSG_ID>" \
  --format json
```

Use the cached token:
```bash
vcert enroll -p ngts -t <ACCESS_TOKEN> -z "<YOUR_ISSUING_TEMPLATE_NAME>" --cn "app.example.com" --no-prompt
```

---

## 3. Verifying Authentication & Credentials

To test credentials and verify that the account has permission to read the issuing template:

**For SaaS (API Key):**
```bash
vcert getpolicy -p vcp -k "<YOUR_SAAS_API_KEY>" -z "<APPLICATION_NAME>\<ISSUING_TEMPLATE>"
```

**For NGTS (OAuth):**
```bash
vcert getpolicy -p ngts \
  --token-url "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token" \
  --client-id "$VCERT_CLIENT_ID" \
  --client-secret "$VCERT_CLIENT_SECRET" \
  --scope "tsg_id:<10_DIGIT_TSG_ID>" \
  -z "<YOUR_ISSUING_TEMPLATE_NAME>"
```

A successful response outputs the JSON certificate policy and CA rules.
