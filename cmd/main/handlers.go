package main

import (
	"log/slog"
	"os"

	"github.com/moonlags/sherstjanka/internal/flux"
)

func (s *server) imageHandler() {
	gen := flux.New(os.Getenv("FAL_KEY"))

	for photo := range s.photos {
		slog.Info("generating image", "photo", photo)

		url, err := gen.GenerateImage(photo.Prompt)
		if err != nil {
			slog.Error("Can not generate image", "err", err)

			msg, err := generationFailure(photo)
			if err != nil {
				slog.Error("Can not get model response to generation failure", "err", err)
				continue
			}

			if _, err := s.bot.Send(msg); err != nil {
				slog.Error("Can not send message", "err", err)
			}
			continue
		}

		msg, err := generationSuccess(photo, url)
		if err != nil {
			slog.Error("Can not get model response to generation success", "err", err)
			continue
		}

		if _, err := s.bot.Send(msg); err != nil {
			slog.Error("Can not send message", "err", err)
		}
	}
}
