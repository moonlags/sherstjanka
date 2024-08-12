package main

import (
	"context"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/moonlags/sherstjanka/internal/photo"
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

	model := newModel(client)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}

	server := server{
		client: client,
		bot:    bot,
		model:  model,
		chats:  make(map[int64]*genai.ChatSession),
		photos: make(chan *photo.Photo, 5),
	}

	server.run()
}
