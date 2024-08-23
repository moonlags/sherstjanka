package chats

import (
	"context"
	"sync"

	"github.com/google/generative-ai-go/genai"
)

type Chat struct {
	*genai.ChatSession
	mu sync.Mutex
}

func NewChat(model *genai.GenerativeModel) *Chat {
	return &Chat{
		ChatSession: model.StartChat(),
	}
}

func (c *Chat) SendAsync(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	c.mu.Lock()

	resp, err := c.SendMessage(ctx, parts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
