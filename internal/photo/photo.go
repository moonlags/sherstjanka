package photo

import "github.com/google/generative-ai-go/genai"

type Photo struct {
	ChatSession *genai.ChatSession
	Prompt      string
	ChatID      int64
	MessageID   int
}
