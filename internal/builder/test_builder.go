package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func BuildTestData(ctx context.Context, inputFile, outputDir, apiKey string, numSamples int) (err error) {
	var inputData json.RawMessage
	if inputData, err = os.ReadFile(inputFile); err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}
	var client = openai.NewClient(option.WithAPIKey(apiKey))
	for i := 0; i < numSamples; i++ {
		if res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.F(openai.ChatModelGPT4o2024_08_06),
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("You are a helpful assistant that generates test data for a JSON schema."),
				openai.UserMessage(fmt.Sprintf("Wherever there is a union type, choose the %d-th option, modulo the number of options.", i)),
			}),
			ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
				openai.ResponseFormatJSONSchemaParam{
					Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
					JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
						Name:   openai.F(fmt.Sprintf("build-test-data-%d", i)),
						Schema: openai.F(any(string(inputData))),
					}),
				},
			),
		}); err != nil {
			return fmt.Errorf("generating test data: %w", err)
		} else {
			var outputData json.RawMessage
			if err := json.Unmarshal([]byte(res.Choices[0].Message.Content), &outputData); err != nil {
				return fmt.Errorf("unmarshalling test data: %w", err)
			}
			if err := os.WriteFile(fmt.Sprintf("%s/%s_data_%d.json", outputDir, fileNameWithoutExtension(inputFile), i), outputData, 0644); err != nil {
				return fmt.Errorf("writing test data: %w", err)
			}
		}
	}
	return nil
}

func fileNameWithoutExtension(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}
