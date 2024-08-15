package http

import (
	"bytes"
	"encoding/json"
	"github.com/v4lli/AIecho/pipeline/internal/config"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strings"
)

type I2TPrompt struct {
	Image       []byte    `json:"image"`
	Temperature int       `json:"temperature"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func RunImage2Text(cfg *config.Config, img *image.Image, maxTokens int) string {

	var buffer bytes.Buffer
	err := jpeg.Encode(&buffer, *img, nil)
	if err != nil {
		log.Printf("Error encoding image for i2t: %v", err)
		return ""
	}
	prompt := I2TPrompt{
		Image:       buffer.Bytes(),
		Temperature: 0,
		Messages: []Message{
			{
				Role:    "system",
				Content: "Provide a detailed comma separated bullet point list of items, people and interactions in the image",
			},
		},
		MaxTokens: maxTokens,
	}
	jsonI2TPrompt, err := json.Marshal(prompt)
	if err != nil {
		log.Printf("Error encoding prompt: %v", err)
		return ""
	}
	url := cfg.GenerateI2TURL()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonI2TPrompt))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return ""
	}
	req.Header = cfg.GenerateHeader()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing request: %v", err)
		return ""
	}
	defer resp.Body.Close()

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		log.Printf("Error decoding response: %v", err)
		return ""
	}
	if resp.StatusCode == 200 {
		result, ok := jsonResponse["result"].(map[string]interface{})
		if !ok {
			log.Printf("Error decoding I2T response: %v", jsonResponse["result"])
			return ""
		}
		description := sanitizeResponse(result["description"].(string))
		return description
	}
	log.Printf("Error response not type 200, %v", jsonResponse)
	return ""
}

func sanitizeResponse(description string) string {
	if index := strings.LastIndex(description, "."); index != -1 {
		description = description[:index+1]
	}
	return description
}
