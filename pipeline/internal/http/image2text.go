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

type Result struct {
	ResType string `json:"type"`
	Content string `json:"content"`
	Urgent  bool   `json:"urgent"`
}

func RunImage2Text(cfg *config.Config, img gocv.Mat, maxTokens int) Result {
	buffer, err := gocv.IMEncode(".png", img)
	if err != nil {
		log.Printf("Error encoding image for i2t: %v", err)
		return Result{
			ResType: "error",
			Content: "Error encoding image for i2t",
			Urgent:  false,
		}
	}
	defer buffer.Close()

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
		return Result{
			ResType: "error",
			Content: "Error encoding prompt",
			Urgent:  false,
		}
	}
	url := cfg.GenerateI2TURL()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonI2TPrompt))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return Result{
			ResType: "error",
			Content: "Error creating request",
			Urgent:  false,
		}
	}
	req.Header = cfg.GenerateHeader()

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Printf("Error executing request: %v", err)
		return Result{
			ResType: "error",
			Content: "Error executing request",
			Urgent:  false,
		}
	}
	defer resp.Body.Close()

	imgErr := img.Close()
	if imgErr != nil {
		return Result{
			ResType: "error",
			Content: "Error closing image",
			Urgent:  false,
		}
	}

	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		log.Printf("Error decoding response: %v", err)
		return Result{
			ResType: "error",
			Content: "Error decoding response",
			Urgent:  false,
		}
	}

	if resp.StatusCode == 200 {
		result, ok := jsonResponse["result"].(map[string]interface{})
		if !ok {
			log.Printf("Error decoding I2T response: %v", jsonResponse["result"])
			return Result{
				ResType: "error",
				Content: "Error decoding I2T response",
				Urgent:  false,
			}
		}
		description := sanitizeResponse(result["description"].(string))
		return Result{
			ResType: "tl",
			Content: description,
			Urgent:  false,
		}
	}
	log.Printf("Error response not type 200, %v", jsonResponse)
	return Result{
		ResType: "error",
		Content: "Error response not type 200",
		Urgent:  false,
	}
}

func sanitizeResponse(description string) string {
	if index := strings.LastIndex(description, "."); index != -1 {
		description = description[:index+1]
	}
	return description
}
