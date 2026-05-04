// Package main demonstrates sending an email using the Anima Go SDK.
//
// Run with:
//
//	ANIMA_API_KEY=ak_live_... go run . --agent-id=agent_xxx --to=user@example.com
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	anima "github.com/anima-labs-ai/go"
)

func main() {
	agentID := flag.String("agent-id", "", "Agent ID to send from")
	to := flag.String("to", "", "Recipient email address")
	subject := flag.String("subject", "Hello from Anima", "Email subject")
	body := flag.String("body", "This email was sent by an Anima agent using the Go SDK.", "Email body")
	flag.Parse()

	if *agentID == "" || *to == "" {
		flag.Usage()
		os.Exit(1)
	}

	apiKey := os.Getenv("ANIMA_API_KEY")
	if apiKey == "" {
		log.Fatal("ANIMA_API_KEY environment variable is required")
	}

	client := anima.NewClient(apiKey)
	ctx := context.Background()

	// Send the email.
	msg, err := client.Messages.SendEmail(ctx, anima.SendEmailParams{
		AgentID: *agentID,
		To:      []string{*to},
		Subject: *subject,
		Body:    *body,
		BodyHTML: fmt.Sprintf("<html><body><p>%s</p></body></html>", *body),
	})
	if err != nil {
		// Demonstrate error handling.
		if errors.Is(err, anima.ErrValidation) {
			log.Fatalf("Invalid request: %v", err)
		}
		if errors.Is(err, anima.ErrAuthentication) {
			log.Fatalf("Authentication failed - check your API key: %v", err)
		}
		if errors.Is(err, anima.ErrRateLimit) {
			var apiErr *anima.APIError
			if errors.As(err, &apiErr) {
				log.Fatalf("Rate limited. Retry after %d seconds.", apiErr.RetryAfter)
			}
		}
		log.Fatalf("Failed to send email: %v", err)
	}

	fmt.Printf("Email sent successfully!\n")
	fmt.Printf("  Message ID: %s\n", msg.ID)
	fmt.Printf("  Status:     %s\n", msg.Status)
	fmt.Printf("  From:       %s\n", msg.FromAddress)
	fmt.Printf("  To:         %s\n", msg.ToAddress)

	// Retrieve the message to confirm.
	retrieved, err := client.Messages.Get(ctx, msg.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve message: %v", err)
	}
	fmt.Printf("  Confirmed:  %s\n", retrieved.Status)
}
