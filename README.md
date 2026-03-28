# Anima Go SDK

Official Go client library for the [Anima](https://anima.com) API.

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
    anima.WithBaseURL("https://api.staging.anima.com"),
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
| `client.Agents` | Create, list, update, delete agents; rotate API keys |
| `client.Organizations` | Manage organizations and master keys |
| `client.Messages` | Send email/SMS, list and search messages |
| `client.Emails` | List emails, manage attachments |
| `client.Domains` | Add/verify domains, DNS records, deliverability stats |
| `client.Cards` | Virtual cards, spending policies, transactions, approvals |
| `client.Phones` | Provision/release phone numbers |
| `client.Vault` | Credential vault: store, search, generate passwords, TOTP |
| `client.Security` | Content scanning, security events |
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
