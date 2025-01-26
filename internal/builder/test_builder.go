package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"
	anthroption "github.com/anthropics/anthropic-sdk-go/option"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	jsonschema "github.com/santhosh-tekuri/jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder/messages"
	"github.com/tylergannon/go-gen-jsonschema/internal/common"
)

func BuildTestDataAnthropic(ctx context.Context, inputFile, outputDir, apiKey string, numSamples int) (err error) {
	var schema *jsonschema.Schema
	var inputData json.RawMessage
	if inputData, err = os.ReadFile(inputFile); err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}
	if schema, err = loadSchema(inputData, inputFile); err != nil {
		return fmt.Errorf("loading schema: %w", err)
	}
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	var wg sync.WaitGroup
	var client = anthropic.NewClient(anthroption.WithAPIKey(apiKey))
	var errs = make([]error, numSamples)
	for i := 0; i < numSamples; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = buildOneAnthropic(ctx, inputData, schema, fileNameWithoutExtension(inputFile), outputDir, i, client, false)
		}(i)
	}
	wg.Wait()
	return errors.Join(errs...)
}

func buildOneAnthropic(ctx context.Context, inputData json.RawMessage, schema *jsonschema.Schema, objectName, outputDir string, i int, client *anthropic.Client, forceCreate bool) (err error) {
	var (
		outputPath = fmt.Sprintf("%s/%s_data_%d.json", outputDir, objectName, i)
		outputData string
	)

	fmt.Printf("Building test data for %s, sample %d\n", objectName, i)
	var sb strings.Builder
	sb.WriteString("You are a helpful assistant that generates test data for a JSON schema.\n")
	sb.WriteString(fmt.Sprintf("Wherever there is a union type, choose the %d-th option, modulo the number of options.\n Meaning if there are 5 options choose the %d%%5-th option (the %dth).\n", i, i, i%5))
	sb.WriteString(fmt.Sprintf("I'll give you the schema.  Please respond with a JSON object that conforms to the schema.\nSurround the JSON object with %s...%s tags.\n", common.JSONTag, common.JSONTagEnd))
	sb.WriteString("The schema is:\n")
	sb.Write(inputData)
	if res, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(common.AnthropicModel),
		MaxTokens: anthropic.F(int64(1024)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(sb.String())),
		}),
	}); err != nil {
		return fmt.Errorf("generating test data: %w", err)
	} else {
		if outputData, err = messages.ExtractJsonResponse(res.Content[0].Text); err != nil {
			return fmt.Errorf("extracting JSON response: %w", err)
		}
		if err := schema.Validate(bytes.NewReader([]byte(outputData))); err != nil {
			return fmt.Errorf("invalid test data: %w", err)
		}
		if err := os.WriteFile(outputPath, []byte(outputData), 0644); err != nil {
			return fmt.Errorf("writing test data: %w", err)
		}
	}
	// Make another request to get assertions.
	return nil
}

func BuildTestDataOpenAI(ctx context.Context, inputFile, outputDir, apiKey string, numSamples int) (err error) {
	var inputData json.RawMessage
	if inputData, err = os.ReadFile(inputFile); err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}
	if err = os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	var client = openai.NewClient(option.WithAPIKey(apiKey))
	for i := 0; i < numSamples; i++ {
		if res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model: openai.F(common.OpenAIModel),
			Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("You are a helpful assistant that generates test data for a JSON schema."),
				openai.UserMessage(fmt.Sprintf("Wherever there is a union type, choose the %d-th option, modulo the number of options.", i)),
			}),
			ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
				openai.ResponseFormatJSONSchemaParam{
					Type: openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
					JSONSchema: openai.F(openai.ResponseFormatJSONSchemaJSONSchemaParam{
						Name:   openai.F(fmt.Sprintf("build-test-data-%d", i)),
						Schema: openai.F[any](inputData),
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

// loadSchema loads and compiles a json schema from the generated json schema
// located at filePath, for use in validating schema prior to unmarshaling.
func loadSchema(rawMessage json.RawMessage, fileName string) (*jsonschema.Schema, error) {
	var (
		compiler = jsonschema.NewCompiler()
		schema   *jsonschema.Schema
		err      error
	)
	if err = compiler.AddResource(fileName, bytes.NewReader(rawMessage)); err != nil {
		return nil, fmt.Errorf("loading schema document %s: %w", fileName, err)
	} else if schema, err = compiler.Compile(fileName); err != nil {
		return nil, fmt.Errorf("compiling schema for file %s: %w", fileName, err)
	}
	return schema, nil
}
