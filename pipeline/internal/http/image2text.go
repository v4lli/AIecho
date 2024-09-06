package http

import (
	"bytes"
	"encoding/json"
	"github.com/v4lli/AIecho/pipeline/internal/config"
	"gocv.io/x/gocv"
	"log"
	"net/http"
	"strings"
)

type I2TPrompt struct {
	Temperature float64   `json:"temperature"`
	Prompt      string    `json:"prompt,omitempty"`
	Raw         bool      `json:"raw,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
	Image       []int     `json:"image"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func RunImage2Text(cfg *config.Config, img gocv.Mat, maxTokens int) string {
	buffer, err := gocv.IMEncode(".png", img)
	//displayImage(img)
	if err != nil {
		log.Printf("Error encoding image for i2t: %v", err)
		return ""
	}
	bufferStorage := buffer.GetBytes()
	intImageArray := make([]int, len(bufferStorage))
	for i, b := range bufferStorage {
		intImageArray[i] = int(b)
	}
	prompt := I2TPrompt{
		Temperature: 0.7,
		Raw:         true,
		Messages: []Message{
			{
				Role:    "system",
				Content: "Provide a detailed comma separated bullet point list of items, people and interactions in the image",
			},
		},
		Image:     intImageArray,
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

func displayImage(mat gocv.Mat) {
	window := gocv.NewWindow("trial")
	defer window.Close()
	window.IMShow(mat)
	window.WaitKey(0)
}
