package main

import (
	"fmt"
	"log"

	"github.com/moonlags/sherstjanka/internal/flux"
)

func (s *server) imageHandler() {
	for photo := range s.photos {
		data, err := flux.GenerateImage(photo.Prompt, true, nil)
		if err != nil {
			fmt.Println(err)

			msg, err := generationFailure(photo)
			if err != nil {
				log.Fatal("Failed to get model response to generation failure:", err)
			}

			if _, err := s.bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			continue
		}

		msg, err := generationSuccess(photo, data)
		if err != nil {
			log.Fatal("Failed to get model response to generation success:", err)
		}

		if _, err := s.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
	}
}
