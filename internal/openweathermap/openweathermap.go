package openweathermap

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Config struct {
	ApiKey string
}

func New(apikey string) *Config {
	return &Config{
		ApiKey: apikey,
	}
}

type Response struct {
	Message string `json:"message"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp float32 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float32 `json:"speed"`
	} `json:"wind"`
}

func (cfg *Config) Weather(city string) (*Response, error) {
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, cfg.ApiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	weather := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(weather); err != nil {
		return nil, err
	}

	if weather.Message != "" {
		return nil, fmt.Errorf("error from openweathermap: %v", err)
	}

	return weather, nil
}
