package aiclient

import (
	"AIGateway/internal/config"
	"context"
	"log"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type Client struct {
	client openai.Client
	cfg    config.AiClient
}

func NewClient(cfg config.AiClient) *Client {
	client := openai.NewClient(option.WithAPIKey(cfg.Key), option.WithBaseURL(cfg.Url))
	return &Client{client: client, cfg: cfg}
}

func (c *Client) Fetch(ctx context.Context, payload string) (*responses.Response, error) {
	//таймаут на один вызов AI
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	prompt := "ты помощник, который кратко резюмирует входящий текст в 1-2 предложениях, без домыслов, только то, что есть в тексте"

	resp, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt + "\n\n" + payload)},
		Model: openai.ChatModelGPT5_2,
	})
	if err != nil {
		log.Printf("fetch response error: %s", err)
		return nil, err
	}
	return resp, nil
}
