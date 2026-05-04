package providers

import (
	"context"

	"github.com/Brownei/aitop/process"
	"github.com/Brownei/aitop/utils"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	apiKey string
	client *genai.Client
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	ctx := context.Background()
	c, _ := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	return &GeminiProvider{
		apiKey: apiKey,
		client: c,
	}
}

func (g *GeminiProvider) Name() string {
	return "Google Gemini"
}

func (g *GeminiProvider) ValidateAPIKey(ctx context.Context) error {
	parts := []*genai.Part{
		{Text: "Hello world, How are you doing?"},
	}
	_, err := g.client.Models.GenerateContent(ctx, "gemini-1.5-flash", []*genai.Content{{Parts: parts}}, nil)
	if err != nil {
		return err
	}

	return nil
}

func (g *GeminiProvider) Analyze(ctx context.Context, p process.ProcessInfo) (string, error) {
	parts := []*genai.Part{
		{Text: utils.GetGlobalPrompt(p)},
	}
	resp, err := g.client.Models.GenerateContent(ctx, "gemini-1.5-flash", []*genai.Content{{Parts: parts}}, nil)
	if err != nil {
		return err.Error(), err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			text := part.Text
			return text, nil
		}
	}

	return "Unable to analyze", nil
}
