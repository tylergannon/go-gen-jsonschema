package typealternatives

import (
	jsonschema "github.com/tylergannon/go-gen-jsonschema"
	"github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/typealternatives/subpkg"
	"time"
)

type (
	LLMFriendlyTime subpkg.LLMFriendlyTime

	TimedSomething struct {
		Time LLMFriendlyTime `json:"time"`
	}
)

func TimeAgoToLLMFriendlyTime(t subpkg.TimeAgo) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(-subpkg.ToDuration(t.Unit, t.Quantity))), nil
}

func FromNowToLLMFriendlyTime(t subpkg.TimeFromNow) (LLMFriendlyTime, error) {
	return LLMFriendlyTime(time.Now().Add(subpkg.ToDuration(t.Unit, t.Value))), nil
}

func ActualTimeToLLMFriendlyTime(t subpkg.ActualTime) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}

func NowToLLMFriendlyTime(t subpkg.Now) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}
func BeginningOfTimeToLLMFriendlyTime(t subpkg.BeginningOfTime) (LLMFriendlyTime, error) {
	_t, err := t.ToTime()
	return LLMFriendlyTime(_t), err
}

var _ = jsonschema.SetTypeAlternative[LLMFriendlyTime](
	// For referencing a time in the past using relative units
	jsonschema.Alt("timeAgo", TimeAgoToLLMFriendlyTime),
	// For referencing a time in the future using relative units
	jsonschema.Alt("timeFromNow", FromNowToLLMFriendlyTime),
	// When given an actual time. Must be valid RFC3339 time format
	jsonschema.Alt("actualTime", ActualTimeToLLMFriendlyTime),
	// To refer to the present moment
	jsonschema.Alt("now", NowToLLMFriendlyTime),
	// To reference all of history
	jsonschema.Alt("beginningOfTime", BeginningOfTimeToLLMFriendlyTime),
)

const JSONSchemaTimedSomething = `{
        "type": "object",
        "properties": {
          "time": {
            "anyOf": [
              {
                "description": "TimeAgo reflects a relative time in the past, given in units of time relative to the present time.\n\n\n\n## **Properties**\n\n### unit\n\nChoose the unit of as given.\n\n### quantity\n\nEnter the number of the selected unit.\n\n",
                "type": "object",
                "properties": {
                  "__type__": {
                    "const": "TimeAgo",
                    "type": "string"
                  },
                  "unit": {
                    "$ref": "#/$defs/TimeUnit"
                  },
                  "quantity": {
                    "type": "integer"
                  }
                },
                "required": [
                  "__type__",
                  "unit",
                  "quantity"
                ],
                "additionalProperties": false
              },
              {
                "description": "\n\n\n\n## **Properties**\n\n### unit\n\nChoose the unit of as given.\n\n### value\n\nEnter the number of the selected unit.\n\n",
                "type": "object",
                "properties": {                                                                                                                                                                         "__type__": {
                    "const": "TimeFromNow",
                    "type": "string"
                  },
                  "unit": {
                    "$ref": "#/$defs/TimeUnit"
                  },
                  "value": {
                    "type": "integer"
                  }
                },
                "required": [
                  "__type__",
                  "unit",
                  "value"
                ],
                "additionalProperties": false
              },
              {
                "type": "object",
                "properties": {
                  "__type__": {
                    "const": "ActualTime",
                    "type": "string"
                  },
                  "dateTime": {
                    "type": "string"
                  }
                },
                "required": [
                  "__type__",
                  "dateTime"
                ],
                "additionalProperties": false
              },
              {
                "type": "object",
                "properties": {
                  "__type__": {
                    "const": "Now",
                    "type": "string"
                  }
                },
                "required": [
                  "__type__"
                ],
                "additionalProperties": false
              },
              {
                "type": "object",
                "properties": {
                  "__type__": {
                    "const": "BeginningOfTime",
                    "type": "string"
                  }
                },
                "required": [
                  "__type__"
                ],
                "additionalProperties": false
              }
            ]
          }
        },
        "required": [
          "time"
        ],
        "additionalProperties": false,
        "$defs": {
          "TimeUnit": {
            "description": "Choose the unit of time given by the user.",
            "enum": [
              "minutes",
              "hours",
              "weeks",
              "days",
              "months",
              "years"
            ],
            "type": "string"
          }
        }
      }`
