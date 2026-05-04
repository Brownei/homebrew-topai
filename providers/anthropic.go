package providers

import (
	"context"

	"github.com/Brownei/aitop/process"
	"github.com/Brownei/aitop/utils"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicProvider struct {
	apiKey string
	client anthropic.Client
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

func (a *AnthropicProvider) Name() string {
	return "Anthropic Claude"
}

func (a *AnthropicProvider) ValidateAPIKey(ctx context.Context) error {
	// Make a simple call to verify the key works
	_, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.ModelClaudeHaiku4_5,
		Messages: []anthropic.MessageParam{
			{
				Role: anthropic.MessageParamRoleUser,
			},
		},
		MaxTokens: 10,
	})
	return err
}

func (a *AnthropicProvider) Analyze(ctx context.Context, p process.ProcessInfo) (string, error) {
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model: anthropic.ModelClaudeHaiku4_5,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(utils.GetGlobalPrompt(p))),
		},
		MaxTokens: 1024,
	})

	if err != nil {
		return "", err
	}

	if len(message.Content) > 0 {
		textBlock := message.Content[0].Text
		return textBlock, nil
	}

	return "Unable to analyze", nil
}
