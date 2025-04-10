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

Your task is to generate one assertion for each field in the test data.

There are three cases:
1. The default case is a value assertion based on the kind of field (string, numeric, bool).
2. When you see a field named "!type", you should use the ResolveTypeInfo tool function to resolve the type information for the field.
   In this case, DO NOT ISSUE A VALUE ASSERTION.  Instead, issue a type assertion that matches the type name returned by the tool function.
3. When you encounter a field that is an array, you should issue an assertion for the length of the array.

Validation requirements:
1. Every "!type" field should have a corresponding type assertion in the final response.
2. Every array field should have a corresponding assertion for the length of the array in the final response.
3. Each pure-value (i.e. not an array or a "!type" field) field should have a corresponding assertion for its value in the final response.
4. Field paths must use dot notation (e.g. ".Field.Child", ".Field.ArrayField.[0].Field)
5. Inability to resolve a type should be ignored silently.
`

const userMessage1 = `
The response must:
1. Be wrapped in '<json>' and '</json>' tags
2. Contain only the JSON object with no additional text or formatting
3. Follow this exact schema:
`

type GetTypeInfoFulfillment func(arg ToolFuncGetTypeInfo) (string, error)

// BuildAssertions calls out to the LLM to get a set of assertions for some test data.
// TestData is the JSON object whose unmarshalling will be the body of the test.
// FlattenedStruct is the Go struct definition that will receive the unmarshalled JSON.
// PkgPath is the package path of the struct definition.
// Client is the Anthropic client to use for the LLM call.
// ToolFunc is the fulfillment function for the tool function used to resolve union types.
func BuildAssertions(ctx context.Context, testData, flattenedStruct, pkgPath string, client *anthropic.Client, toolFunc GetTypeInfoFulfillment) (GeneratedTestResponse, error) {
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
			Model:     common.AnthropicModel,
			MaxTokens: int64(1024),
			System: []anthropic.TextBlockParam{
				{Text: systemMessage},
				{Text: userMessage1 + string(GeneratedTestResponse{}.Schema())},
			},
			Messages: userMessages,
			Tools: []anthropic.ToolUnionParam{
				{
					OfTool: &anthropic.ToolParam{
						Name:        "ResolveTypeInfo",
						Description: anthropic.Opt("Resolve the type information for one or more union type instances found in the test data."),
						InputSchema: anthropic.ToolInputSchemaParam{
							Properties: ToolFuncGetTypeInfo{}.Schema(),
						},
					},
				},
			},
		}); err != nil {
			return GeneratedTestResponse{}, fmt.Errorf("building assertions: %w", err)
		} else if res.StopReason == anthropic.MessageStopReasonToolUse {
			foo, _ := json.MarshalIndent(res.Content, "", "  ")
			fmt.Println("Tool use", string(foo))
			var toolResults []anthropic.ContentBlockParamUnion
			for _, data := range res.Content {
				if data.Type == "tool_use" {
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
