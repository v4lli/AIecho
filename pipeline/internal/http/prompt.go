package http

import (
	"bytes"
	"encoding/json"
	"github.com/v4lli/AIecho/pipeline/internal/config"
	"log"
	"net/http"
)

type LLMPrompt struct {
	Temperature int       `json:"temperature"`
	Messages    []Message `json:"messages"`
}

var llmStorage []Message

func RunLLM(cfg *config.Config, i2tResponses []string, movement bool) Result {
	prompt := LLMPrompt{
		Temperature: 0,
		Messages: []Message{
			{
				Role: "system",
				Content: `input: - 3 consecutive images taken in the same scene, separator: ;
							output: combined single sentence scene description for visually impaired people,
									maximum length 50 words. 
							output content: enumerate individual objects, people and their visual descriptions 
							rules:  don't repeat prompt when response should be continued, 
									don't repeat old response 
									if no information to add just say no more new information 
									use natural language, no control information`,
			},
		},
	}
	for _, i2t := range i2tResponses {
		prompt.Messages = append(
			prompt.Messages, Message{
				Role:    "user",
				Content: i2t,
			},
		)
	}
	if movement {
		llmStorage = []Message{}
	} else {
		prompt.Messages = append(
			prompt.Messages, Message{
				Role:    "user",
				Content: "Please tell me more about the scene without repeating what you already said here:",
			},
		)
		for _, msg := range llmStorage {
			prompt.Messages = append(prompt.Messages, msg)
		}
	}
	jsonLLMPrompt, err := json.Marshal(prompt)
	if err != nil {
		log.Printf("Error marshalling LLM prompt: %v", err)
		return Result{
			ResType: "error",
			Content: "Error marshalling LLM prompt",
			Urgent:  false,
		}
	}
	url := cfg.GenerateLLMURL()
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonLLMPrompt))
	if err != nil {
		log.Printf("Error creating LLM request: %v", err)
		return Result{
			ResType: "error",
			Content: "Error creating LLM request",
			Urgent:  false,
		}
	}
	req.Header = cfg.GenerateHeader()
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Printf("Error creating LLM request: %v", err)
		return Result{
			ResType: "error",
			Content: "Error creating LLM request",
			Urgent:  false,
		}
	}
	var jsonResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	if err != nil {
		log.Printf("Error unmarshalling LLM response: %v", err)
		return Result{
			ResType: "error",
			Content: "Error unmarshalling LLM response",
			Urgent:  false,
		}
	}
	if resp.StatusCode == 200 {
		result, ok := jsonResponse["result"].(map[string]interface{})
		if !ok {
			log.Printf("Error decoding LLM response: %v", jsonResponse["result"])
			return Result{
				ResType: "error",
				Content: "Error decoding LLM response",
				Urgent:  false,
			}
		}
		description := sanitizeResponse(result["response"].(string))
		llmStorage = append(
			llmStorage, Message{
				Role:    "assistant",
				Content: description,
			},
		)
		return Result{
			ResType: "desc",
			Content: description,
			Urgent:  false,
		}
	}
	return Result{
		ResType: "error",
		Content: "Error response not type 200",
		Urgent:  false,
	}
}
