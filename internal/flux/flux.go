package flux

import (
	"net/http"
)

type Config struct {
	ApiKey string
}

func New(key string) *Config {
	return &Config{
		ApiKey: key,
	}
}

func (c *Config) GenerateImage(prompt string) (string, error) {
	req, err := http.NewRequest("POST", "https://fal.run/fal-ai/flux/schnell", promptToJson(prompt))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+c.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	url, err := parseImageURL(resp.Body)
	if err != nil {
		return "", err
	}

	return url, nil
}
