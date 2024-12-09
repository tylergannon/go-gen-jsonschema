package structs

import "github.com/tylergannon/go-gen-jsonschema/internal/builder/testfixtures/structs/enums"

// StructWithBasicTypes is a struct with basic types
type StructWithBasicTypes struct {
	// Foo is an int
	Foo int `json:"foo"`
	// Bar is a string
	Bar int `json:"bar"`

	// Baz is a slice of int
	Baz []int `json:"baz"`
}

// StructWithInline is a struct with inline structs for fields
type StructWithInline struct {
	// Foo is blah
	Foo int `json:"foo"`
	// Bar hates hot dogs
	Bar int `json:"bar"`
	// Baz doesn't do "shower"
	Baz []int `json:"baz"`
	// Quux is French
	Quux struct {
		// Foo is funny
		Foo int `json:"foo"`
		// Bar is a dog
		Bar int `json:"bar"`
		// Baz was a robot for halloween
		Baz  []int `json:"baz"`
		Quux struct {
			// Foo will too!!
			Foo int `json:"foo"`
			// Bar be que
			Bar int `json:"bar"`
			// Baz really?
			Baz []int `json:"baz"`
		} `json:"quux99"`
	} `json:"__quux__"`
}

// StructWithNamedTypes has some definitions in it.
type StructWithNamedTypes struct {
	// Foo is very Foo
	Foo StructWithInline `json:"foo"`
	// Bar is ABCDE
	Bar StructWithInline `json:"bar"`
	// Baz is very cookie
	Baz []StructWithInline `json:"baz"`
}

// Hotdog is a delectable treat that is enjoyed by Americans and lovers of
// American cuisine the whole world over.
type Hotdog struct {
	// The length of the dog in inches
	LengthInches int `json:"lengthInches"`
	// Choose your favorite toppings
	Toppings []struct {
		// Choose from the available items
		Item enums.HotdogTopping `json:"item"`
		// how many scoops? (more if you really like it)
		Scoops int `json:"scoops"`
	} `json:"toppings"`

	// what kind of dog will you choose?
	Type enums.HotdogType `json:"type"`
}

const StructWithBasicTypesSchema = `
{
    "description": "StructWithBasicTypes is a struct with basic types",
    "type": "object",
    "properties": {
        "foo": {
            "description": "Foo is an int",
            "type": "integer"
        },
        "bar": {
            "description": "Bar is a string",
            "type": "integer"
        },
        "baz": {
            "description": "Baz is a slice of int",
            "items": {
                "type": "integer"
            },
            "type": "array"
        }
    },
    "required": [
        "foo",
        "bar",
        "baz"
    ],
    "additionalProperties": false
}`

const StructWithInlineSchema = `{
    "description": "StructWithInline is a struct with inline structs for fields",
    "type": "object",
    "properties": {
        "foo": {
            "description": "Foo is blah",
            "type": "integer"
        },
        "bar": {
            "description": "Bar hates hot dogs",
            "type": "integer"
        },
        "baz": {
            "description": "Baz doesn't do \"shower\"",
            "items": {
                "type": "integer"
            },
            "type": "array"
        },
        "__quux__": {
            "description": "Quux is French",
            "type": "object",
            "properties": {
                "foo": {
                    "description": "Foo is funny",
                    "type": "integer"
                },
                "bar": {
                    "description": "Bar is a dog",
                    "type": "integer"
                },
                "baz": {
                    "description": "Baz was a robot for halloween",
                    "items": {
                        "type": "integer"
                    },
                    "type": "array"
                },
                "quux99": {
                    "type": "object",
                    "properties": {
                        "foo": {
                            "description": "Foo will too!!",
                            "type": "integer"
                        },
                        "bar": {
                            "description": "Bar be que",
                            "type": "integer"
                        },
                        "baz": {
                            "description": "Baz really?",
                            "items": {
                                "type": "integer"
                            },
                            "type": "array"
                        }
                    },
                    "required": [
                        "foo",
                        "bar",
                        "baz"
                    ],
                    "additionalProperties": false
                }
            },
            "required": [
                "foo",
                "bar",
                "baz",
                "quux99"
            ],
            "additionalProperties": false
        }
    },
    "required": [
        "foo",
        "bar",
        "baz",
        "__quux__"
    ],
    "additionalProperties": false
}`

const StructWithNamedTypesSchema = `{
    "description": "StructWithNamedTypes has some definitions in it.\n\n\n\n## **Properties**\n\n### foo\n\nis very Foo\n\n### bar\n\nis ABCDE\n\n### baz\n\nis very cookie\n\n",
    "type": "object",
    "properties": {
        "foo": {
            "$ref": "#/$defs/StructWithInline"
        },
        "bar": {
            "$ref": "#/$defs/StructWithInline"
        },
        "baz": {
            "items": {
                "$ref": "#/$defs/StructWithInline"
            },
            "type": "array"
        }
    },
    "required": [
        "foo",
        "bar",
        "baz"
    ],
    "additionalProperties": false,
    "$defs": {
        "StructWithInline": {
            "description": "StructWithInline is a struct with inline structs for fields",
            "type": "object",
            "properties": {
                "foo": {
                    "description": "Foo is blah",
                    "type": "integer"
                },
                "bar": {
                    "description": "Bar hates hot dogs",
                    "type": "integer"
                },
                "baz": {
                    "description": "Baz doesn't do \"shower\"",
                    "items": {
                        "type": "integer"
                    },
                    "type": "array"
                },
                "__quux__": {
                    "description": "Quux is French",
                    "type": "object",
                    "properties": {
                        "foo": {
                            "description": "Foo is funny",
                            "type": "integer"
                        },
                        "bar": {
                            "description": "Bar is a dog",
                            "type": "integer"
                        },
                        "baz": {
                            "description": "Baz was a robot for halloween",
                            "items": {
                                "type": "integer"
                            },
                            "type": "array"
                        },
                        "quux99": {
                            "type": "object",
                            "properties": {
                                "foo": {
                                    "description": "Foo will too!!",
                                    "type": "integer"
                                },
                                "bar": {
                                    "description": "Bar be que",
                                    "type": "integer"
                                },
                                "baz": {
                                    "description": "Baz really?",
                                    "items": {
                                        "type": "integer"
                                    },
                                    "type": "array"
                                }
                            },
                            "required": [
                                "foo",
                                "bar",
                                "baz"
                            ],
                            "additionalProperties": false
                        }
                    },
                    "required": [
                        "foo",
                        "bar",
                        "baz",
                        "quux99"
                    ],
                    "additionalProperties": false
                }
            },
            "required": [
                "foo",
                "bar",
                "baz",
                "__quux__"
            ],
            "additionalProperties": false
        }
    }
}`

const HotDogSchema = `{
    "description": "Hotdog is a delectable treat that is enjoyed by Americans and lovers of American cuisine the whole world over.",
    "type": "object",
    "properties": {
        "lengthInches": {
            "description": "The length of the dog in inches",
            "type": "integer"
        },
        "toppings": {
            "description": "Choose your favorite toppings",
            "items": {
                "description": "\n\n## **Properties**\n\n### item\n\nChoose from the available items\n\n### scoops\n\nhow many scoops? (more if you really like it)\n\n",
                "type": "object",
                "properties": {
                    "item": {
                        "description": "HotdogTopping is a string for hotdog toppings\n\n## Values\n\n### mustard\n\nBold and tangy zing.\n\n### ketchup\n\nSweet and bright.\n\n### mayonnaise\n\nRich and creamy.\n\n### relish\n\nCrisp, pickled crunch.\n\n### sauerkraut\n\nSharp, fermented bite.",
                        "enum": [
                            "mustard",
                            "ketchup",
                            "mayonnaise",
                            "relish",
                            "sauerkraut"
                        ],
                        "type": "string"
                    },
                    "scoops": {
                        "type": "integer"
                    }
                },
                "required": [
                    "item",
                    "scoops"
                ],
                "additionalProperties": false
            },
            "type": "array"
        },
        "type": {
            "description": "what kind of dog will you choose?",
            "enum": [
                "beef",
                "chicken",
                "turkey",
                "veggie"
            ],
            "type": "string"
        }
    },
    "required": [
        "lengthInches",
        "toppings",
        "type"
    ],
    "additionalProperties": false
}`
