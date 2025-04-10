package common

import (
	"io"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go"
)

const (
	AnthropicModel = anthropic.ModelClaude3_7SonnetLatest
	OpenAIModel    = openai.ChatModelGPT4oMini2024_07_18
	JSONTag        = "<json>"
	JSONTagEnd     = "</json>"
)

func LogClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Printf("failed to close output file: %v", err)
	}
}
