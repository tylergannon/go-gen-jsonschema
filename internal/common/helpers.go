package common

import (
	"io"
	"log"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go"
)

const (
	AnthropicModel = anthropic.ModelClaude3_5SonnetLatest
	OpenAIModel    = openai.ChatModelGPT4o2024_08_06
	JSONTag        = "<json>"
	JSONTagEnd     = "</json>"
)

func LogClose(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Printf("failed to close output file: %v", err)
	}
}
