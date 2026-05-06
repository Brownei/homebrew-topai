package providers

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	apiKey string
	client *openai.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		client: openai.NewClient(apiKey),
	}
}

func (o *OpenAIProvider) Name() string {
	return "OpenAI GPT-4"
}

func (o *OpenAIProvider) ValidateAPIKey(ctx context.Context) error {
	_, err := o.client.ListModels(ctx)
	return err
}

func (o *OpenAIProvider) Analyze(ctx context.Context, content string) (string, error) {

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT5ChatLatest,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "Unable to analyze", nil
}
