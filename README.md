# Anima Go SDK

Official Go client library for the [Anima](https://useanima.sh) API.

## Installation

```bash
go get github.com/anima-labs/anima-go
```

Requires Go 1.22 or later.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    anima "github.com/anima-labs/anima-go"
)

func main() {
    client := anima.NewClient("ak_live_your_api_key")
    ctx := context.Background()

    // Create an agent.
    agent, err := client.Agents.Create(ctx, anima.CreateAgentParams{
        OrgID: "org_123",
        Name:  "My Agent",
        Slug:  "my-agent",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created agent: %s\n", agent.ID)

    // Send an email.
    msg, err := client.Messages.SendEmail(ctx, anima.SendEmailParams{
        AgentID: agent.ID,
        To:      []string{"user@example.com"},
        Subject: "Hello from Anima",
        Body:    "Sent by an AI agent.",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Sent message: %s (status: %s)\n", msg.ID, msg.Status)
}
```

## Configuration

Use functional options to customize the client:

```go
import "time"

client := anima.NewClient("ak_live_...",
    anima.WithBaseURL("https://api.staging.useanima.sh"),
    anima.WithTimeout(10 * time.Second),
    anima.WithMaxRetries(5),
)
```

### Custom HTTP Client

```go
import "net/http"

httpClient := &http.Client{
    Timeout: 15 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns: 100,
    },
}

client := anima.NewClient("ak_live_...",
    anima.WithHTTPClient(httpClient),
)
```

## Error Handling

The SDK uses sentinel errors with `errors.Is` and typed errors with `errors.As`:

```go
import "errors"

// Check error category with errors.Is
_, err := doSomething()
if errors.Is(err, anima.ErrNotFound) {
    fmt.Println("Resource not found")
} else if errors.Is(err, anima.ErrRateLimit) {
    fmt.Println("Rate limited, try again later")
} else if errors.Is(err, anima.ErrAuthentication) {
    fmt.Println("Invalid API key")
} else if errors.Is(err, anima.ErrValidation) {
    fmt.Println("Invalid request parameters")
}

// Extract detailed error info with errors.As
var apiErr *anima.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d\n", apiErr.Status)
    fmt.Printf("Code: %s\n", apiErr.Code)
    fmt.Printf("Message: %s\n", apiErr.Message)

    if apiErr.RetryAfter > 0 {
        fmt.Printf("Retry after %d seconds\n", apiErr.RetryAfter)
    }
}
```

### Available Sentinel Errors

| Error | Description |
|-------|-------------|
| `ErrAuthentication` | Invalid or missing API key (401/403) |
| `ErrNotFound` | Resource not found (404) |
| `ErrValidation` | Invalid request parameters (400/422) |
| `ErrRateLimit` | Rate limit exceeded (429) |
| `ErrConflict` | Resource conflict (409) |
| `ErrInternalServer` | Server error (5xx) |
| `ErrTimeout` | Request timed out |
| `ErrNetwork` | Network-level error |
| `ErrRetryExhausted` | All retry attempts failed |

## Resource Services

All resource services are available as fields on the `Client`:

| Service | Description |
|---------|-------------|
| `client.A2A` | Agent-to-agent task submission, listing, and cancellation |
| `client.Addresses` | Physical/mailing addresses: CRUD, validation |
| `client.Agents` | Create, list, update, delete agents; rotate API keys |
| `client.Anomaly` | Anomaly alerts, detection rules, baselines, quarantine |
| `client.Audit` | Immutable audit logs: list, get, export (CSV/JSON) |
| `client.Cards` | Virtual cards, spending policies, transactions, approvals |
| `client.Compliance` | Controls, reports, dashboard, DSARs (SOC2/GDPR/PCI) |
| `client.Domains` | Add/verify domains, DNS records, deliverability stats |
| `client.Emails` | List emails, manage attachments |
| `client.Identity` | DID documents, key rotation, verifiable credentials, agent cards |
| `client.Messages` | Send email/SMS, list and search messages |
| `client.Organizations` | Manage organizations and master keys |
| `client.Phones` | Provision/release phone numbers |
| `client.Pods` | Compute pods: create, list, update, delete, usage stats |
| `client.Registry` | Public agent registry: register, search, lookup, update, unlist |
| `client.Security` | Content scanning, security events |
| `client.Vault` | Credential vault: store, search, generate passwords, TOTP |
| `client.Wallet` | Crypto wallets: create, pay, X-402 fetch, transactions, freeze |
| `client.Webhooks` | Webhook CRUD, test delivery, list deliveries |

### Sending Email

```go
msg, err := client.Messages.SendEmail(ctx, anima.SendEmailParams{
    AgentID: "agent_123",
    To:      []string{"user@example.com"},
    Subject: "Hello",
    Body:    "Plain text body",
    BodyHTML: "<h1>Hello</h1>",
})
```

### Managing Cards

```go
card, err := client.Cards.Create(ctx, anima.CreateCardParams{
    AgentID:         "agent_123",
    Label:           "Marketing Budget",
    SpendLimitDaily: anima.IntPtr(10000), // $100.00
})

// Freeze a card instantly.
card, err = client.Cards.Freeze(ctx, card.ID)

// Kill switch: freeze all cards for an agent.
result, err := client.Cards.KillSwitch(ctx, anima.KillSwitchParams{
    AgentID: "agent_123",
    Active:  true,
})
```

### Agent Identity (DID)

```go
// Get an agent's DID document.
did, err := client.Identity.GetDID(ctx, "agent_123")
fmt.Printf("DID: %s\n", did.ID)

// Verify a credential.
result, err := client.Identity.VerifyCredential(ctx, "eyJhbGciOi...")
fmt.Printf("Valid: %v\n", result.Valid)
```

### Agent Registry

```go
// Register an agent in the public registry.
entry, err := client.Registry.Register(ctx, anima.RegisterAgentParams{
    DID:  "did:anima:agent_123",
    Name: "My Agent",
})

// Search the registry.
results, err := client.Registry.Search(ctx, anima.RegistrySearchParams{
    Query: "email assistant",
})
```

### Wallet

```go
// Create a wallet for an agent.
wallet, err := client.Wallet.Create(ctx, "agent_123", nil)

// Make a payment.
payment, err := client.Wallet.Pay(ctx, "agent_123", anima.WalletPayParams{
    To:     "did:anima:recipient",
    Amount: "1.50",
})

// Freeze a wallet.
err = client.Wallet.Freeze(ctx, "agent_123")
```

### Pods

```go
// Create a compute pod.
pod, err := client.Pods.Create(ctx, anima.CreatePodParams{
    AgentID: "agent_123",
    Name:    "worker-1",
    Image:   "agent-runtime:latest",
})

// Check usage.
usage, err := client.Pods.Usage(ctx, pod.ID)
fmt.Printf("CPU: %.2f%%\n", usage.CPU)
```

### A2A (Agent-to-Agent)

```go
// Submit a task to another agent.
task, err := client.A2A.SubmitTask(ctx, "agent_456", anima.SubmitA2ATaskParams{
    Input: map[string]string{"query": "summarize this document"},
})

// Check task status.
task, err = client.A2A.GetTask(ctx, "agent_456", task.ID)
fmt.Printf("Status: %s\n", task.Status)
```

### Addresses

```go
// Create an address for an agent.
addr, err := client.Addresses.Create(ctx, anima.CreateAddressParams{
    AgentID:    "agent_123",
    Line1:      "123 Main St",
    City:       "San Francisco",
    State:      "CA",
    PostalCode: "94105",
    Country:    "US",
})

// Validate the address.
validation, err := client.Addresses.Validate(ctx, addr.ID, "agent_123")
fmt.Printf("Deliverable: %v\n", validation.Deliverable)
```

### Vault Credentials

```go
cred, err := client.Vault.CreateCredential(ctx, anima.CreateVaultCredentialParams{
    AgentID: "agent_123",
    Type:    anima.CredentialTypeLogin,
    Name:    "GitHub",
    Login: &anima.VaultLoginData{
        Username: "bot@company.com",
        Password: "s3cr3t",
    },
})
```

### Audit Logs

```go
// List audit logs with filters.
page, err := client.Audit.List(ctx, "org_123", &anima.AuditLogListParams{
    ActorType:    anima.AuditActorAgent,
    ResourceType: "message",
    From:         "2026-01-01T00:00:00Z",
    To:           "2026-03-01T00:00:00Z",
})

// Export audit logs as CSV.
export, err := client.Audit.Export(ctx, "org_123", &anima.AuditLogExportParams{
    Format: anima.AuditExportFormatCSV,
    From:   "2026-01-01T00:00:00Z",
})
fmt.Printf("Download: %s (%d records)\n", export.URL, export.RecordCount)
```

### Compliance

```go
// Seed SOC 2 controls for an organization.
seed, err := client.Compliance.SeedFramework(ctx, "org_123", anima.SeedFrameworkInput{
    Framework: anima.ComplianceFrameworkSOC2,
})
fmt.Printf("Created %d controls\n", seed.ControlsCreated)

// View compliance dashboard.
dashboard, err := client.Compliance.GetDashboard(ctx, "org_123")
fmt.Printf("Overall score: %.1f%%\n", dashboard.OverallScore)

// Create a GDPR data subject access request.
dsar, err := client.Compliance.CreateDSAR(ctx, "org_123", anima.CreateDSARInput{
    SubjectEmail: "user@example.com",
    RequestType:  anima.DSARRequestTypeDeletion,
})
```

### Anomaly Detection

```go
// Create an anomaly detection rule.
rule, err := client.Anomaly.CreateRule(ctx, "org_123", anima.CreateAnomalyRuleInput{
    Name:      "High email volume",
    Metric:    anima.AnomalyMetricEmailSendRate,
    Condition: anima.AnomalyConditionZScoreGT,
    Threshold: 3.0,
    Severity:  anima.AnomalySeverityCritical,
})

// List triggered alerts.
alerts, err := client.Anomaly.ListAlerts(ctx, "org_123", &anima.AnomalyAlertListParams{
    Status: anima.AnomalyAlertStatusTriggered,
})

// Quarantine a misbehaving agent.
q, err := client.Anomaly.Quarantine(ctx, "org_123", "agent_123", anima.QuarantineInput{
    Level:  anima.QuarantineLevelHard,
    Reason: "Unusual email send volume",
})
```

## Pagination

List endpoints return paginated results. Use `ListAutoPaging` for automatic iteration:

```go
iter := client.Agents.ListAutoPaging(&anima.AgentListParams{
    ListParams: anima.ListParams{Limit: 50},
    OrgID:      "org_123",
})

for iter.Next(ctx) {
    agent := iter.Current()
    fmt.Println(agent.Name)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}
```

Or fetch a single page manually:

```go
page, err := client.Agents.List(ctx, &anima.AgentListParams{
    ListParams: anima.ListParams{Limit: 10},
})
for _, agent := range page.Items {
    fmt.Println(agent.Name)
}
```

## Webhook Verification

Verify incoming webhook signatures using HMAC-SHA256:

```go
import "net/http"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    payload, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("Anima-Signature")

    event, err := anima.ConstructWebhookEvent(payload, signature, "whsec_your_secret", nil)
    if err != nil {
        http.Error(w, "Invalid signature", http.StatusForbidden)
        return
    }

    switch event.Type {
    case "agent.created":
        fmt.Printf("New agent: %v\n", event.Data["name"])
    case "message.received":
        fmt.Printf("Message from: %v\n", event.Data["fromAddress"])
    }

    w.WriteHeader(http.StatusOK)
}
```

## Automatic Retries

The SDK automatically retries failed requests with exponential backoff:

- **Retryable status codes:** 429 (Rate Limited), 5xx (Server Errors)
- **Non-retryable:** 400, 401, 403, 404, 409 (returned immediately)
- **Backoff:** Exponential with jitter (1s, 2s, 4s base delays)
- **Retry-After:** Honored when present in response headers
- **Default retries:** 3 (configurable with `WithMaxRetries`)

## License

See [LICENSE](LICENSE) for details.
