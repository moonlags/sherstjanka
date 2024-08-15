package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/moonlags/sherstjanka/internal/chats"
	"github.com/moonlags/sherstjanka/internal/flux"
	"google.golang.org/api/option"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load .env file:", err)
	}
}

func main() {
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		slog.Error("Can not create gemini client", "err", err)
		os.Exit(1)
	}
	defer client.Close()

	model := newModel(client)

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		slog.Error("Can not create telegram bot", "err", err)
		os.Exit(1)
	}

	whitelist, err := strconv.ParseInt(os.Getenv("WHITELIST"), 10, 64)
	if err != nil {
		slog.Error("Can not get whilelist id", "err", err)
		os.Exit(1)
	}

	server := server{
		client:    client,
		bot:       bot,
		model:     model,
		chats:     chats.New(),
		image:     flux.New(os.Getenv("FAL_KEY")),
		whitelist: whitelist,
	}

	server.run()
}
