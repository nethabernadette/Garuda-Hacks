package ai

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultGroqBaseURL = "https://api.groq.com/openai/v1"
	defaultGroqModel   = "llama-3.1-8b-instant"
	defaultTimeout     = 20 * time.Second
)

// Config contains AI provider settings loaded from environment variables.
type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	Enabled     bool
	Timeout     time.Duration
	Temperature float64
}

// LoadConfigFromEnv loads Groq configuration with local-safe defaults.
func LoadConfigFromEnv() Config {
	model := strings.TrimSpace(os.Getenv("GROQ_MODEL"))
	if model == "" {
		model = defaultGroqModel
	}

	baseURL := strings.TrimSpace(os.Getenv("GROQ_BASE_URL"))
	if baseURL == "" {
		baseURL = defaultGroqBaseURL
	}

	enabled := true
	if raw := strings.TrimSpace(os.Getenv("ENABLE_AI")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		enabled = err == nil && parsed
	}

	return Config{
		APIKey:      strings.TrimSpace(os.Getenv("GROQ_API_KEY")),
		Model:       model,
		BaseURL:     strings.TrimRight(baseURL, "/"),
		Enabled:     enabled,
		Timeout:     defaultTimeout,
		Temperature: 0.1,
	}
}
