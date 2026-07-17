package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GroqClient calls Groq's OpenAI-compatible chat completion API.
type GroqClient struct {
	config Config
	client *http.Client
}

// NewGroqClient creates a reusable Groq client.
func NewGroqClient(config Config) *GroqClient {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &GroqClient{
		config: config,
		client: &http.Client{Timeout: timeout},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// CompleteJSON asks Groq for JSON-only output and unmarshals it into dst.
func (c *GroqClient) CompleteJSON(ctx context.Context, systemPrompt string, payload interface{}, dst interface{}) error {
	if !c.config.Enabled {
		return ErrAIUnavailable
	}
	if strings.TrimSpace(c.config.APIKey) == "" {
		log.Println("groq request skipped: missing GROQ_API_KEY")
		return ErrAIUnavailable
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqBody := chatCompletionRequest{
		Model: c.config.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt + "\nReturn valid JSON only. Do not wrap JSON in markdown code fences."},
			{Role: "user", Content: string(body)},
		},
		Temperature: c.config.Temperature,
	}
	rawReq, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, c.config.BaseURL+"/chat/completions", bytes.NewReader(rawReq))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("groq request failed: %v", err)
		return ErrAIUnavailable
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("groq non-2xx response: status=%d body=%s", resp.StatusCode, safeSnippet(respBody))
		return ErrAIUnavailable
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		log.Printf("groq response decode failed: %v", err)
		return ErrInvalidAIResponse
	}
	if len(completion.Choices) == 0 || strings.TrimSpace(completion.Choices[0].Message.Content) == "" {
		log.Println("groq response was empty")
		return ErrInvalidAIResponse
	}

	content := cleanJSONContent(completion.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), dst); err != nil {
		log.Printf("groq json output invalid: %v", err)
		return ErrInvalidAIResponse
	}
	return nil
}

func cleanJSONContent(content string) string {
	cleaned := strings.TrimSpace(content)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```JSON")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimSuffix(cleaned, "```")
	return strings.TrimSpace(cleaned)
}

func safeSnippet(body []byte) string {
	const max = 300
	text := string(body)
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) > max {
		text = text[:max] + "..."
	}
	return fmt.Sprintf("%q", text)
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
