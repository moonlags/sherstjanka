package flux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type generationInfo struct {
	Prompt            string
	RandomiseSeed     bool
	Seed              int
	Width             int
	Height            int
	NumInferenceSteps int
}

func GenerateImage(prompt string, randomiseSeed bool, opts *GenerationOptions) ([]byte, error) {
	info := optsToInfo(prompt, randomiseSeed, opts)

	eventID, err := info.getEventID()
	if err != nil {
		return nil, err
	}

	fmt.Println(eventID)

	resp, err := http.Get("https://black-forest-labs-flux-1-schnell.hf.space/call/infer/" + eventID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	url, err := parseImageURL(resp.Body)
	if err != nil {
		return nil, err
	}

	return getImageBytes(url)
}

func (info *generationInfo) json() *bytes.Reader {
	data := fmt.Sprintf("{\"data\":[\"%s\",%d,%t,%d,%d,%d]}", info.Prompt, info.Seed, info.RandomiseSeed, info.Width, info.Height, info.NumInferenceSteps)

	return bytes.NewReader([]byte(data))
}

func (info *generationInfo) getEventID() (string, error) {
	resp, err := http.Post("https://black-forest-labs-flux-1-schnell.hf.space/call/infer", "application/json", info.json())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var eventID struct {
		EventID string `json:"event_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&eventID); err != nil {
		return "", err
	}

	return eventID.EventID, nil
}
