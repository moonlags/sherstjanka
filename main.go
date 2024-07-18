package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load .env file:", err)
	}
}

func main() {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatal("Failed to create gemini client:", err)
	}
	defer client.Close()
	instructions, err := os.ReadFile("instructions.txt")
	if err != nil {
		log.Fatal("Failed to read instructions from instructions.txt:", err)
	}
	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(string(instructions))},
	}
	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockOnlyHigh,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockOnlyHigh,
		},
	}
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}
	server := NewServer(client, bot, model)
	server.Run(tgbotapi.NewUpdate(0))
}
