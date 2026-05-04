// Package main demonstrates basic usage of the Anima Go SDK.
//
// Run with:
//
//	ANIMA_API_KEY=ak_live_... go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	anima "github.com/anima-labs-ai/go"
)

func main() {
	apiKey := os.Getenv("ANIMA_API_KEY")
	if apiKey == "" {
		log.Fatal("ANIMA_API_KEY environment variable is required")
	}

	client := anima.NewClient(apiKey)
	ctx := context.Background()

	// Create an agent.
	agent, err := client.Agents.Create(ctx, anima.CreateAgentParams{
		OrgID: "org_example",
		Name:  "My First Agent",
		Slug:  "my-first-agent",
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	fmt.Printf("Created agent: %s (ID: %s)\n", agent.Name, agent.ID)

	// List all agents with auto-pagination.
	iter := client.Agents.ListAutoPaging(nil)
	for iter.Next(ctx) {
		a := iter.Current()
		fmt.Printf("  - %s (%s)\n", a.Name, a.Status)
	}
	if err := iter.Err(); err != nil {
		log.Fatalf("Failed to list agents: %v", err)
	}

	// List domains.
	domains, err := client.Domains.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list domains: %v", err)
	}
	fmt.Printf("\nDomains (%d):\n", len(domains.Items))
	for _, d := range domains.Items {
		fmt.Printf("  - %s (verified: %v)\n", d.Domain, d.Verified)
	}

	// List webhooks.
	webhooks, err := client.Webhooks.List(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list webhooks: %v", err)
	}
	fmt.Printf("\nWebhooks (%d):\n", len(webhooks.Items))
	for _, w := range webhooks.Items {
		fmt.Printf("  - %s (active: %v, events: %v)\n", w.URL, w.Active, w.Events)
	}
}
