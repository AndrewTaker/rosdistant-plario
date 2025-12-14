package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

var (
	ErrLimitReached = fmt.Errorf("today's limit for model expired")
)

type Groq struct {
	Token        string
	Model        Model
	Instructions string
	logger       *slog.Logger
}

func NewGroq(token string, model Model, instructions string, logger *slog.Logger) *Groq {
	return &Groq{
		Token:        token,
		Model:        model,
		Instructions: instructions,
		logger:       logger,
	}
}

func (g *Groq) SendGroqRequest(client *http.Client, question string) (*GroqResponse, error) {
	reqBody := GroqRequest{
		Model:            string(g.Model),
		IncludeReasoning: false,
		Messages: []Message{
			{Role: "system", Content: g.Instructions},
			{Role: "user", Content: question},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrLimitReached
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gr GroqResponse
	if err := json.Unmarshal(respBody, &gr); err != nil {
		return nil, fmt.Errorf("err unmarshaling %s", string(respBody))
	}

	return &gr, nil
}

type GroqRequest struct {
	Model            string    `json:"model"`
	IncludeReasoning bool      `json:"include_reasoning"`
	Messages         []Message `json:"messages"`
}

type GroqResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message,omitempty"`
	Text         string  `json:"text,omitempty"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
