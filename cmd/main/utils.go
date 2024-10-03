package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/generative-ai-go/genai"
	"github.com/moonlags/sherstjanka/internal/openweathermap"
)

func (s *server) parseReponse(update tgbotapi.Update, response genai.Part) string {
	slog.Info("parsing response", "response", response)

	funcall, ok := response.(genai.FunctionCall)
	if !ok {
		return fmt.Sprint(response)
	}

	s.parseFuncall(update, funcall)
	return ""
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

func (s *server) parseFuncall(update tgbotapi.Update, funcall genai.FunctionCall) {
	switch funcall.Name {
	case modelTools().FunctionDeclarations[0].Name:
		prompt, err := getImageGenerationPrompt(funcall)
		if err != nil {
			slog.Error("Can not get image generation prompt")
			return
		}

		slog.Info("generating image", "prompt", prompt)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		image, err := s.image.GenerateImage(prompt)
		if err != nil {
			slog.Error("Can not generate image:", "err", err)
			s.generationFailure(ctx, update, prompt, err)
		}

		s.generationSuccess(ctx, update, prompt, image)
	case modelTools().FunctionDeclarations[1].Name:
		city, err := getWeatherCity(funcall)
		if err != nil {
			slog.Error("Can not get city for weather", "err", err)
			return
		}

		slog.Info("Getting weather", "city", city)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		weather, err := s.weather.Weather(city)
		if err != nil {
			slog.Error("Can not get weather", "city", city, "err", err)
			s.weatherFail(ctx, update, city, err)
		}

		s.weatherSuccess(ctx, update, city, weather)
	default:
		slog.Error("Unknown function call", "name", funcall.Name)
	}
}

func (s *server) weatherFail(ctx context.Context, update tgbotapi.Update, city string, err error) {
	apiResult := map[string]any{
		"city":  city,
		"error": err,
	}

	parts, err := s.chats.Send(ctx, update.FromChat().ID, genai.FunctionResponse{
		Name:     modelTools().FunctionDeclarations[1].Name,
		Response: apiResult,
	})
	if err != nil {
		slog.Error("Can not get model response", "err", err)
		return
	}

	slog.Warn("Model response to weather fail", "parts", parts)

	for _, part := range parts {
		response := s.parseReponse(update, part)

		if response == "" {
			continue
		}

		msg := tgbotapi.NewMessage(update.FromChat().ID, response)
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := s.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}

func (s *server) generationFailure(ctx context.Context, update tgbotapi.Update, prompt string, err error) {
	apiResult := map[string]any{
		"error":  err,
		"prompt": prompt,
	}

	parts, err := s.chats.Send(ctx, update.FromChat().ID, genai.FunctionResponse{
		Name:     modelTools().FunctionDeclarations[0].Name,
		Response: apiResult,
	})
	if err != nil {
		slog.Error("Can not get model response", "err", err)
		return
	}

	slog.Warn("Model response to generation failure", "parts", parts)

	for _, part := range parts {
		response := s.parseReponse(update, part)

		if response == "" {
			continue
		}

		msg := tgbotapi.NewMessage(update.FromChat().ID, response)
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := s.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}

func (s *server) weatherSuccess(ctx context.Context, update tgbotapi.Update, city string, weather *openweathermap.Response) {
	slog.Info("weather success", "city", city, "weather", weather)

	apiResult := map[string]any{
		"city":        city,
		"temperature": weather.Main.Temp,
		"description": weather.Weather[0].Description,
		"wind":        weather.Wind.Speed,
	}

	parts, err := s.chats.Send(ctx, update.FromChat().ID, genai.FunctionResponse{
		Name:     modelTools().FunctionDeclarations[1].Name,
		Response: apiResult,
	})
	if err != nil {
		slog.Error("Can not get model response to weather success", "err", err)
		return
	}

	slog.Info("Model response to weather success", "parts", parts)

	for _, part := range parts {
		response := s.parseReponse(update, part)

		if response == "" {
			continue
		}

		msg := tgbotapi.NewMessage(update.FromChat().ID, response)
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := s.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}

func (s *server) generationSuccess(ctx context.Context, update tgbotapi.Update, prompt string, image []byte) {
	apiResult := map[string]any{
		"image":  "image is ready",
		"prompt": prompt,
	}

	parts, err := s.chats.Send(ctx, update.FromChat().ID, genai.ImageData("png", image), genai.FunctionResponse{
		Name:     modelTools().FunctionDeclarations[0].Name,
		Response: apiResult,
	})
	if err != nil {
		slog.Error("Can not get model response to generation success", "err", err)

		msg := tgbotapi.NewPhoto(update.FromChat().ID, tgbotapi.FileBytes{Bytes: image})
		msg.ReplyToMessageID = update.Message.MessageID

		if _, err := s.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
			return
		}
	}

	slog.Info("Model response to generation success", "parts", parts)

	for _, part := range parts {
		response := s.parseReponse(update, part)

		if response == "" {
			continue
		}

		msg := tgbotapi.NewPhoto(update.FromChat().ID, tgbotapi.FileBytes{Bytes: image})
		msg.ReplyToMessageID = update.Message.MessageID
		msg.Caption = response

		if _, err := s.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}
