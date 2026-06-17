package main

import (
	"context"
	"fmt"
	"log"

	server "multi-agent/server"

	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	agents, err := server.NewAgentsServer(ctx, "./config.yaml")
	if err != nil {
		log.Fatalf("Failed to initialize AgentsServer: %v", err)
	}

	sessionID := uuid.New().String()
	_, err = agents.SessionService.Create(ctx, &session.CreateRequest{
		UserID:    "demo_user",
		SessionID: sessionID,
		AppName:   "ecommerce",
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	userMsg := genai.NewContentFromText("Bên bạn có những loại sản phẩm gì thế?", genai.RoleUser)

	fmt.Printf("Agent: ")
	for event, err := range agents.Runner.Run(ctx, "demo_user", sessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			log.Fatalf("\nRun error: %v", err)
		}
		if event.Content != nil && len(event.Content.Parts) > 0 {
			fmt.Print(event.Content.Parts[0].Text)
		}
	}
	fmt.Println()
}
