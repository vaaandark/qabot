package chatter

import "github.com/vaaandark/qabot/pkg/chatcontext"

type CompletionRequest struct {
	Model    string                `json:"model"`
	Messages []chatcontext.Message `json:"messages"`
	Stream   bool                  `json:"stream"`
}

func CompletionRequestFromContext(model string, messages []chatcontext.Message) CompletionRequest {
	return CompletionRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}
}

type CompletionResponse struct {
	Choices []Choice `json:"choices"`
}

func (cr CompletionResponse) GetMessage() *chatcontext.Message {
	if len(cr.Choices) == 0 {
		return nil
	}
	return &cr.Choices[0].Message
}

type Choice struct {
	Index   int                 `json:"index"`
	Message chatcontext.Message `json:"message"`
}
