package main

import (
	"context"
	"fmt"

	anima "github.com/anima-labs-ai/go"
)

func main() {
	c := anima.NewClient("")
	ctx := context.Background()

	// Test 1: List voices
	voices, err := c.Voices.List(ctx, anima.ListVoicesParams{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Voices: %d\n", len(voices.Voices))
	if len(voices.Voices) > 0 {
		v := voices.Voices[0]
		fmt.Printf("First: %s | %s | %s | %s | %s\n", v.ID, v.Name, v.Provider, v.Tier, v.Gender)
	}

	// Test 2: List calls
	calls, err := c.Calls.List(ctx, anima.ListCallsParams{})
	if err != nil {
		fmt.Printf("List calls error: %v\n", err)
	} else {
		fmt.Printf("\nCalls: %d (total: %d)\n", len(calls.Calls), calls.Total)
		for _, call := range calls.Calls {
			fmt.Printf("  Call: %s | %s | %s | %s -> %s | tier=%s\n",
				call.ID, call.Direction, call.State, call.From, call.To, call.Tier)
		}
	}

	// Test 3: Get specific call
	if len(calls.Calls) > 0 {
		callID := calls.Calls[0].ID
		call, err := c.Calls.Get(ctx, callID)
		if err != nil {
			fmt.Printf("Get call error: %v\n", err)
		} else {
			fmt.Printf("\nGet call %s:\n", callID)
			fmt.Printf("  Direction: %s\n  State: %s\n  From: %s\n  To: %s\n  Tier: %s\n",
				call.Direction, call.State, call.From, call.To, call.Tier)
		}
	}

	// Test 4: Get transcript (expect empty for unanswered call)
	if len(calls.Calls) > 0 {
		callID := calls.Calls[0].ID
		transcript, err := c.Calls.GetTranscript(ctx, callID)
		if err != nil {
			fmt.Printf("\nTranscript for %s: error (expected for unanswered call): %v\n", callID, err)
		} else {
			fmt.Printf("\nTranscript for %s: %d segments\n", callID, len(transcript.Segments))
		}
	}
}
