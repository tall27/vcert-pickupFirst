# NGTS Credential Configuration & Run Guide

This guide explains from the very beginning how to configure, store, and use credentials to run **VCert** against **CyberArk Certificate Manager SaaS / Palo Alto Networks NGTS**.

---

## 1. Required NGTS Parameters

Before configuring credentials, obtain these 5 values from your NGTS service account portal:

| Parameter | Flag | Environment Variable | Example |
|---|---|---|---|
| **API Base URL** | `-u, --url` | `VCERT_URL` | `https://api.strata.paloaltonetworks.com/ngts` |
| **Token URL** | `--token-url` | `VCERT_TOKEN_URL` | `https://auth.apps.paloaltonetworks.com/am/oauth2/access_token` |
| **Client ID** | `--client-id` | `VCERT_CLIENT_ID` | `your-service-account-client-id` |
| **Client Secret** | `--client-secret` | `VCERT_CLIENT_SECRET` | `your-service-account-client-secret` |
| **Scope (TSG ID)** | `--scope` | `VCERT_SCOPE` | `tsg_id:<10_DIGIT_TSG_ID>` |
| **Issuing Template (Zone)** | `-z, --zone` | `VCERT_ZONE` | `<YOUR_ISSUING_TEMPLATE_NAME>` |

> [!NOTE]
> NGTS uses OAuth 2.0 Client Credentials. The `--scope` value must always follow the format `tsg_id:<10-digit-id>`, where `<10-digit-id>` is your Tenant Service Group identifier.

---

## 2. Credential Configuration Methods

Choose the configuration method that best fits your workflow:

### Method A: Environment Variables

Setting environment variables allows VCert CLI and playbooks to authenticate automatically without passing secrets on the command line.

#### Windows (PowerShell):
```powershell
$env:VCERT_PLATFORM      = "ngts"
$env:VCERT_URL           = "https://api.strata.paloaltonetworks.com/ngts"
$env:VCERT_TOKEN_URL     = "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token"
$env:VCERT_CLIENT_ID     = "your-service-account-client-id"
$env:VCERT_CLIENT_SECRET = "your-service-account-client-secret"
$env:VCERT_SCOPE         = "tsg_id:<10_DIGIT_TSG_ID>"
$env:VCERT_ZONE          = "<YOUR_ISSUING_TEMPLATE_NAME>"
```

#### Linux / macOS (Bash):
```bash
export VCERT_PLATFORM="ngts"
export VCERT_URL="https://api.strata.paloaltonetworks.com/ngts"
export VCERT_TOKEN_URL="https://auth.apps.paloaltonetworks.com/am/oauth2/access_token"
export VCERT_CLIENT_ID="your-service-account-client-id"
export VCERT_CLIENT_SECRET="your-service-account-client-secret"
export VCERT_SCOPE="tsg_id:<10_DIGIT_TSG_ID>"
export VCERT_ZONE="<YOUR_ISSUING_TEMPLATE_NAME>"
```

---

### Method B: Configuration File (`vcert.ini`)

You can store connection details in an INI file (e.g., `vcert.ini`):

```ini
[default]
platform = ngts
url = https://api.strata.paloaltonetworks.com/ngts
ngts_token_url = https://auth.apps.paloaltonetworks.com/am/oauth2/access_token
ngts_client_id = your-service-account-client-id
ngts_client_secret = your-service-account-client-secret
ngts_scope = tsg_id:<10_DIGIT_TSG_ID>
```

To run commands with the INI file:
```bash
vcert enroll --config ./vcert.ini -z "<YOUR_ISSUING_TEMPLATE_NAME>" --cn app.example.com --no-prompt
```

---

### Method C: Playbook YAML Configuration

For automated certificate deployment and cluster convergence (`pickupFirst`), store connection parameters in the `config.connection` section of your playbook YAML. You can reference environment variables with `${VAR_NAME}`:

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
      pickupFirst: true                 # Automatic multi-node convergence
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

For **CyberArk Certificate Manager, SaaS** (`platform: vcp`) authenticating with an API key:

```yaml
config:
  connection:
    platform: vcp
    credentials:
      apiKey: "${VCP_APIKEY}"

certificateTasks:
  - name: shared-service-cert
    renewBefore: 5d
    request:
      pickupFirst: true                 # Automatic multi-node convergence
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

---

### Method D: OAuth Access Token Caching with `getcred`

If you prefer to obtain a short-lived bearer token once and reuse it across multiple commands:

#### 1. Generate the access token:
```bash
vcert getcred -p ngts \
  --token-url "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token" \
  --client-id "your-client-id" \
  --client-secret "your-client-secret" \
  --scope "tsg_id:<10_DIGIT_TSG_ID>" \
  --format json
```

#### 2. Use the token via `-t` or `VCERT_TOKEN`:
```bash
vcert enroll -p ngts -t eyJhbGciOi... -z "<YOUR_ISSUING_TEMPLATE_NAME>" --cn "app.example.com" --no-prompt
```

---

## 3. Verifying Authentication & Credentials

To verify that your credentials work and that the service account has permission to read the issuing template, run `getpolicy`:

```bash
# When using environment variables:
vcert getpolicy

# Or passing flags explicitly:
vcert getpolicy -p ngts \
  --token-url "https://auth.apps.paloaltonetworks.com/am/oauth2/access_token" \
  --client-id "$VCERT_CLIENT_ID" \
  --client-secret "$VCERT_CLIENT_SECRET" \
  --scope "tsg_id:<10_DIGIT_TSG_ID>" \
  -z "<YOUR_ISSUING_TEMPLATE_NAME>"
```

A successful response outputs the JSON certificate policy and issuing rules.

---

## 4. Running Certificate Operations

Once credentials are configured (e.g. via environment variables):

### Enroll a Certificate:
```bash
vcert enroll \
  -z "<YOUR_ISSUING_TEMPLATE_NAME>" \
  --cn "web1.example.com" \
  --san-dns "web1.example.com" \
  --cert-file "./cert.crt" \
  --key-file "./cert.key" \
  --chain-file "./chain.crt" \
  --no-prompt
```

### Renew an Existing Certificate:
```bash
vcert renew \
  --thumbprint "file:./cert.crt" \
  --cert-file "./cert_renewed.crt" \
  --key-file "./cert_renewed.key" \
  --chain-file "./chain_renewed.crt" \
  --no-prompt
```

### Run Playbook with `pickupFirst`:
```bash
vcert run -f ./playbook.yaml
```


