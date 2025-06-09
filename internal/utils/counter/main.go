package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// measureSchema recursively crawls a JSON Schema–like object and accumulates
// the length of all property names, definition names, enum values, and const values.
func measureSchema(schema any) int {
	var total int

	switch node := schema.(type) {
	case map[string]any:
		// Check for "properties"
		if props, ok := node["properties"].(map[string]any); ok {
			for propName, propVal := range props {
				total += len(propName)
				total += measureSchema(propVal)
			}
		}

		// Check for "definitions"
		if defs, ok := node["definitions"].(map[string]any); ok {
			for defName, defVal := range defs {
				total += len(defName)
				total += measureSchema(defVal)
			}
		}

		// Check for "enum"
		if enumVals, ok := node["enum"].([]any); ok {
			for _, e := range enumVals {
				if strVal, ok := e.(string); ok {
					total += len(strVal)
				}
			}
		}

		// Check for "const"
		if constVal, ok := node["const"].(string); ok {
			total += len(constVal)
		}

		// Recursively measure anything else that might contain nested schemas
		// (e.g., "items", "allOf", "anyOf", "oneOf", "not", etc.)
		// We exclude keys we've already handled.
		for key, val := range node {
			if key == "properties" || key == "definitions" || key == "enum" || key == "const" {
				continue
			}
			total += measureSchema(val)
		}

	case []any:
		// If it's an array, recursively check each element
		for _, elem := range node {
			total += measureSchema(elem)
		}
	}

	return total
}

func main() {
	// If a file path argument is provided, read from the file;
	// otherwise, read from stdin.
	var data []byte
	var err error

	if len(os.Args) > 1 {
		filePath := os.Args[1]
		data, err = os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Error reading file: %v", err)
		}
	} else {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("Error reading stdin: %v", err)
		}
	}

	// Parse the JSON
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	// Locate response_format.json_schema.schema
	// If missing, print "Structured Response Length: 0"
	responseFormat, ok := root["response_format"].(map[string]any)
	if !ok {
		fmt.Println("Structured Response Length: 0")
	} else {
		jsonSchema, ok := responseFormat["json_schema"].(map[string]any)
		if !ok {
			fmt.Println("Structured Response Length: 0")
		} else {
			schemaObj, ok := jsonSchema["schema"].(map[string]any)
			if !ok {
				fmt.Println("Structured Response Length: 0")
			} else {
				schemaLen := measureSchema(schemaObj)
				fmt.Printf("Structured Response Length: %d\n", schemaLen)
			}
		}
	}

	// Now, check if there are tool functions (often under "functions").
	// If so, measure each function's schema (in "parameters") the same way.
	if functionsVal, ok := root["functions"].([]any); ok {
		for i, fn := range functionsVal {
			fnObj, ok := fn.(map[string]any)
			if !ok {
				continue
			}

			// If you store the JSON schema for the tool function
			// in fnObj["parameters"], measure it:
			if parameters, ok := fnObj["parameters"].(map[string]any); ok {
				length := measureSchema(parameters)
				fmt.Printf("Tool Function [%d] Schema Length: %d\n", i, length)
			}
		}
	}
}
