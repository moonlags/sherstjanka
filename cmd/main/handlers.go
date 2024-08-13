package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moonlags/sherstjanka/internal/flux"
)

func (s *server) imageHandler() {
	gen := flux.New(os.Getenv("FAL_KEY"))

	for photo := range s.photos {
		fmt.Printf("generating image %#v\n", photo)

		url, err := gen.GenerateImage(photo.Prompt)
		if err != nil {
			fmt.Println(err)

			msg, err := generationFailure(photo)
			if err != nil {
				fmt.Println("Failed to get model response to generation failure:", err)
				continue
			}

			if _, err := s.bot.Send(msg); err != nil {
				log.Fatal("Failed to send a message:", err)
			}
			continue
		}

		msg, err := generationSuccess(photo, url)
		if err != nil {
			fmt.Println("Failed to get model response to generation success:", err)
			continue
		}

		if _, err := s.bot.Send(msg); err != nil {
			log.Fatal("Failed to send a message:", err)
		}
	}
}
