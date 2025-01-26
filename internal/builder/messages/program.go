package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/tylergannon/go-gen-jsonschema/internal/common"
)

const systemMessage = `
You are a Go programmer tasked with designing a unit test to validate
custom JSON unmarshaling code. You will be given:

1. A JSON object ("test data")
2. A Go file containing the definition of a struct that will receive the unmarshalled JSON.

The struct definition is "flattened" in the sense that the named types have been
replaced with their underlying types. The exception is when there is a union type,
which is represented as a field with a named type. When you find a named type (that's not
a basic Go type), note that the corresponding entry in the JSON will contain a discriminator
field '!type', used to select the right Go type to instantiate. The '!type' and the named-type
field indicates a union type, where the '!type' field indicates the chosen shape/type for this
instance of the union type.

Use the ResolveTypeInfo tool function to resolve all union types found in the test data.
If the function call returns a union type, you can call it again with the new union
type as you find them. If a union type cannot be resolved, return an error explaining why.

After all tool function calls have been made, the goal is to generate a set of assertions
that validate the test data. This should be returned as a JSON object with a set of assertions.
Note that you're not returning Go code, just field paths and values.

Validation requirements:
1. Each value in the test data must have exactly one assertion
2. For union types: first assert the type, then assert the field values
3. Field paths must use dot notation (e.g. ".parent.child")
4. For array fields: first assert the array length, then make assertions about elements
5. Missing required fields in the test data should be reported as errors
6. Invalid values should be reported as errors with clear explanations
`

const userMessage1 = `
The response must:
1. Be wrapped in '<json>' and '</json>' tags
2. Contain only the JSON object with no additional text or formatting
3. Follow this exact schema:
`

// BuildAssertions calls out to the LLM to get a set of assertions for some test data.
func BuildAssertions(ctx context.Context, testData, flattenedStruct, pkgPath string, client *anthropic.Client, toolFunc func(arg ToolFuncGetTypeInfo) (string, error)) (GeneratedTestResponse, error) {
	var userMessages = []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(testData)),
		anthropic.NewAssistantMessage(anthropic.NewTextBlock("I see the JSON object. I will process the Go struct definition, identify all union types, resolve them using the ResolveTypeInfo tool, and generate assertions for all fields including array lengths. I'll ensure each union type has both type and value assertions.")),
		anthropic.NewUserMessage(anthropic.NewTextBlock(flattenedStruct)),
	}
	var (
		done     bool
		response string
	)
	for !done {
		if res, err := client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.F(common.AnthropicModel),
			MaxTokens: anthropic.F(int64(1024)),
			System: anthropic.F([]anthropic.TextBlockParam{
				anthropic.NewTextBlock(systemMessage),
				anthropic.NewTextBlock(userMessage1 + string(GeneratedTestResponse{}.Schema())),
			}),
			Messages: anthropic.F(userMessages),
			Tools: anthropic.F([]anthropic.ToolParam{
				{
					Name:        anthropic.F("ResolveTypeInfo"),
					Description: anthropic.F("Resolve the type information for one or more union type instances found in the test data."),
					InputSchema: anthropic.F[any](ToolFuncGetTypeInfo{}.Schema()),
				},
			}),
		}); err != nil {
			return GeneratedTestResponse{}, fmt.Errorf("building assertions: %w", err)
		} else if res.StopReason == anthropic.MessageStopReasonToolUse {
			var toolResults []anthropic.ContentBlockParamUnion
			for _, data := range res.Content {
				if data.Type == anthropic.ContentBlockTypeToolUse {
					switch data.Name {
					case "ResolveTypeInfo":
						var arg ToolFuncGetTypeInfo
						if err := json.Unmarshal([]byte(data.Input), &arg); err != nil {
							return GeneratedTestResponse{}, fmt.Errorf("unmarshalling tool call argument: %w", err)
						}
						if result, err := toolFunc(arg); err != nil {
							toolResults = append(toolResults, anthropic.NewToolResultBlock(data.ID, err.Error(), true))
						} else {
							toolResults = append(toolResults, anthropic.NewToolResultBlock(data.ID, result, false))
						}
					}
				}
			}
			userMessages = append(userMessages, res.ToParam())
			userMessages = append(userMessages, anthropic.NewUserMessage(toolResults...))
		} else {
			done = true
			response = res.Content[0].Text
		}
	}

	var result GeneratedTestResponse
	if jsonData, err := ExtractJsonResponse(response); err != nil {
		return GeneratedTestResponse{}, fmt.Errorf("extracting JSON response: %w", err)
	} else if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return GeneratedTestResponse{}, fmt.Errorf("unmarshalling JSON response: %w", err)
	}

	return result, nil
}

func ExtractJsonResponse(input string) (string, error) {
	startPos := strings.Index(input, common.JSONTag)
	endPos := strings.LastIndex(input, common.JSONTagEnd)
	if startPos == -1 || endPos == -1 {
		return "", fmt.Errorf("no %s...%s tags found in response", common.JSONTag, common.JSONTagEnd)
	}
	return input[startPos+len(common.JSONTag) : endPos], nil
}
