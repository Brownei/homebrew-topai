package providers

import (
	"context"
	"flag"

	"google.golang.org/genai"
)

var model = flag.String("model", "gemini-2.5-flash", "the model name, e.g. gemini-2.5-flash")
var config *genai.GenerateContentConfig = &genai.GenerateContentConfig{Temperature: genai.Ptr[float32](0.5)}

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
	chat, err := g.client.Chats.Create(ctx, *model, config, nil)
	if err != nil {
		return err
	}

	// Send first chat message.
	_, err = chat.SendMessage(ctx, genai.Part{Text: "Hello world, How are you doing?"})
	if err != nil {
		return err
	}

	return nil
}

func (g *GeminiProvider) Analyze(ctx context.Context, content string) (string, error) {
	chat, err := g.client.Chats.Create(ctx, *model, config, nil)
	if err != nil {
		return "Unable to create chat", err
	}

	// Send first chat message.
	response, err := chat.SendMessage(ctx, genai.Part{Text: content})
	if err != nil {
		return "Unable to send message", err
	}

	return response.Text(), nil
}
