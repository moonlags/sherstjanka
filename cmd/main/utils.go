package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
)

func (s *server) parseReponse(update tgbotapi.Update, response genai.Part) (tgbotapi.Chattable, error) {
	slog.Info("parsing response", "response", response)

	funcall, ok := response.(genai.FunctionCall)
	if !ok {
		msg := tgbotapi.NewMessage(update.FromChat().ID, fmt.Sprint(response))
		msg.ReplyToMessageID = update.Message.MessageID

		return msg, nil
	}

	prompt, err := getPrompt(funcall)
	if err != nil {
		return nil, err
	}

	slog.Info("generating image", "prompt", prompt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	url, err := s.image.GenerateImage(prompt)
	if err != nil {
		slog.Error("Can not generate image:", "err", err)
		return s.generationFailure(ctx, update, prompt, err)
	}

	return s.generationSuccess(ctx, update, prompt, url)
}

func (s *server) checkWhitelist(update tgbotapi.Update) bool {
	_, err := s.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: s.whitelist,
			UserID: update.Message.From.ID,
		},
	})

	return err == nil
}

func getPrompt(funcall genai.FunctionCall) (string, error) {
	if funcall.Name != imageGenerationTool().FunctionDeclarations[0].Name {
		return "", fmt.Errorf("unknown function call: %v", funcall.Name)
	}

	promptraw, ok := funcall.Args["prompt"]
	if !ok {
		return "", fmt.Errorf("argument prompt not found")
	}

	prompt, ok := promptraw.(string)
	if !ok {
		return "", fmt.Errorf("expected prompt type string got %T", promptraw)
	}

	return prompt, nil
}

func (s *server) generationFailure(ctx context.Context, update tgbotapi.Update, prompt string, err error) (tgbotapi.Chattable, error) {
	apiResult := map[string]any{
		"error":  err,
		"prompt": prompt,
	}

	parts, err := s.chats.Send(ctx, update.FromChat().ID, genai.FunctionResponse{
		Name:     imageGenerationTool().FunctionDeclarations[0].Name,
		Response: apiResult,
	})
	if err != nil {
		return nil, err
	}

	parsed := fmt.Sprint(parts[0])

	msg := tgbotapi.NewMessage(update.FromChat().ID, parsed)
	msg.ReplyToMessageID = update.Message.MessageID

	return msg, nil
}

func (s *server) generationSuccess(ctx context.Context, update tgbotapi.Update, prompt string, url string) (tgbotapi.Chattable, error) {
	slog.Info("generation success", "prompt", prompt, "url", url)

	apiResult := map[string]any{
		"message": "image is ready",
		"prompt":  prompt,
	}

	parts, err := s.chats.Send(ctx, update.FromChat().ID, genai.FunctionResponse{
		Name:     imageGenerationTool().FunctionDeclarations[0].Name,
		Response: apiResult,
	})
	if err != nil {
		slog.Error("Can not get model response to generation success", "err", err)

		msg := tgbotapi.NewPhoto(update.FromChat().ID, tgbotapi.FileURL(url))
		msg.ReplyToMessageID = update.Message.MessageID

		return msg, nil
	}

	parsed := fmt.Sprint(parts[0])

	msg := tgbotapi.NewPhoto(update.FromChat().ID, tgbotapi.FileURL(url))
	msg.ReplyToMessageID = update.Message.MessageID
	msg.Caption = parsed

	return msg, nil
}
