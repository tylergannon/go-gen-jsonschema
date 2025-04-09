package jsonschema

import (
	"context"
	"encoding/json"
	"fmt"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error)
}

// type ToolOption struct {
// 	name              string
// 	description       string
// 	rawParameters     json.RawMessage
// 	paramDescriptions map[string]string
// }

// func WithToolName(name string) ToolOption {
// 	return ToolOption{name: name}
// }

// func WithToolDescription(description string) ToolOption {
// 	return ToolOption{description: description}
// }

// func WithToolParameters(parameters json.RawMessage) ToolOption {
// 	return ToolOption{rawParameters: parameters}
// }

// func WithParamDesc(desc map[string]string) ToolOption {
// 	return ToolOption{paramDescriptions: desc}
// }

func BuildTool(fn any) Tool {
	panic("not implemented")
}

type toolType1[T any] struct {
	toolFn      func(ctx context.Context, arg T) (json.RawMessage, error)
	name        string
	description string
	inputSchema json.RawMessage
}

// Description implements Tool.
func (b *toolType1[T]) Description() string {
	return b.description
}

// Execute implements Tool.
func (b *toolType1[T]) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var arg T
	if err := json.Unmarshal(params, &arg); err != nil {
		return nil, err
	}
	return b.toolFn(ctx, arg)
}

// Name implements Tool.
func (b *toolType1[T]) Name() string {
	return b.name
}

// Parameters implements Tool.
func (b *toolType1[T]) Parameters() json.RawMessage {
	return b.inputSchema
}

var _ Tool = &toolType1[any]{}

func NewTool[T any](name string, description string, inputSchema json.RawMessage, toolFn func(ctx context.Context, arg T) (json.RawMessage, error)) *toolType1[T] {
	return &toolType1[T]{
		name:        name,
		description: description,
		inputSchema: inputSchema,
		toolFn:      toolFn,
	}
}

type toolType2[T, U any] struct {
	toolFn             func(ctx context.Context, arg1 T, arg2 U) (json.RawMessage, error)
	name               string
	arg1Name, arg2Name string
	description        string
	inputSchema        json.RawMessage
}

var _ Tool = &toolType2[any, any]{}

func NewTool2[T, U any](name, description, arg1Name, arg2Name string, inputSchema json.RawMessage, toolFn func(ctx context.Context, arg1 T, arg2 U) (json.RawMessage, error)) *toolType2[T, U] {
	return &toolType2[T, U]{
		name:        name,
		arg1Name:    arg1Name,
		arg2Name:    arg2Name,
		description: description,
		inputSchema: inputSchema,
		toolFn:      toolFn,
	}
}

// Name implements Tool.
func (b *toolType2[T, U]) Name() string {
	return b.name
}

// Parameters implements Tool.
func (b *toolType2[T, U]) Parameters() json.RawMessage {
	return b.inputSchema
}

// Description implements Tool.
func (b *toolType2[T, U]) Description() string {
	return b.description
}

func (b *toolType2[T, U]) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var (
		arg     map[string]json.RawMessage
		arg1raw json.RawMessage
		arg2raw json.RawMessage
		arg1    T
		arg2    U
		ok      bool
		err     error
	)
	if err = json.Unmarshal(params, &arg); err != nil {
		return nil, err
	}
	if arg1raw, ok = arg[b.arg1Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg1Name)
	} else if err = json.Unmarshal(arg1raw, &arg1); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg1Name, err)
	}
	if arg2raw, ok = arg[b.arg2Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg2Name)
	} else if err = json.Unmarshal(arg2raw, &arg2); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg2Name, err)
	}
	return b.toolFn(ctx, arg1, arg2)
}

type toolType3[T, U, V any] struct {
	toolFn                       func(ctx context.Context, arg1 T, arg2 U, arg3 V) (json.RawMessage, error)
	name                         string
	arg1Name, arg2Name, arg3Name string
	description                  string
	inputSchema                  json.RawMessage
}

var _ Tool = &toolType3[any, any, any]{}

func NewTool3[T, U, V any](name, description, arg1Name, arg2Name, arg3Name string, inputSchema json.RawMessage, toolFn func(ctx context.Context, arg1 T, arg2 U, arg3 V) (json.RawMessage, error)) *toolType3[T, U, V] {
	return &toolType3[T, U, V]{
		name:        name,
		arg1Name:    arg1Name,
		arg2Name:    arg2Name,
		arg3Name:    arg3Name,
		description: description,
		inputSchema: inputSchema,
		toolFn:      toolFn,
	}
}

// Description implements Tool.
func (b *toolType3[T, U, V]) Description() string {
	return b.description
}

// Execute implements Tool.
func (b *toolType3[T, U, V]) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var (
		arg     map[string]json.RawMessage
		arg1raw json.RawMessage
		arg2raw json.RawMessage
		arg3raw json.RawMessage
		arg1    T
		arg2    U
		arg3    V
		ok      bool
		err     error
	)
	if err = json.Unmarshal(params, &arg); err != nil {
		return nil, err
	}
	if arg1raw, ok = arg[b.arg1Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg1Name)
	} else if err = json.Unmarshal(arg1raw, &arg1); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg1Name, err)
	}
	if arg2raw, ok = arg[b.arg2Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg2Name)
	} else if err = json.Unmarshal(arg2raw, &arg2); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg2Name, err)
	}
	if arg3raw, ok = arg[b.arg3Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg3Name)
	} else if err = json.Unmarshal(arg3raw, &arg3); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg3Name, err)
	}
	return b.toolFn(ctx, arg1, arg2, arg3)
}

// Name implements Tool.
func (b *toolType3[T, U, V]) Name() string {
	return b.name
}

// Parameters implements Tool.
func (b *toolType3[T, U, V]) Parameters() json.RawMessage {
	return b.inputSchema
}

type toolType4[T, U, V, W any] struct {
	toolFn                                 func(ctx context.Context, arg1 T, arg2 U, arg3 V, arg4 W) (json.RawMessage, error)
	name                                   string
	arg1Name, arg2Name, arg3Name, arg4Name string
	description                            string
	inputSchema                            json.RawMessage
}

var _ Tool = &toolType4[any, any, any, any]{}

func NewTool4[T, U, V, W any](name, description, arg1Name, arg2Name, arg3Name, arg4Name string, inputSchema json.RawMessage, toolFn func(ctx context.Context, arg1 T, arg2 U, arg3 V, arg4 W) (json.RawMessage, error)) *toolType4[T, U, V, W] {
	return &toolType4[T, U, V, W]{
		name:        name,
		arg1Name:    arg1Name,
		arg2Name:    arg2Name,
		arg3Name:    arg3Name,
		arg4Name:    arg4Name,
		description: description,
		inputSchema: inputSchema,
		toolFn:      toolFn,
	}
}

// Description implements Tool.
func (b *toolType4[T, U, V, W]) Description() string {
	return b.description
}

// Execute implements Tool.
func (b *toolType4[T, U, V, W]) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var (
		arg     map[string]json.RawMessage
		arg1raw json.RawMessage
		arg2raw json.RawMessage
		arg3raw json.RawMessage
		arg4raw json.RawMessage
		arg1    T
		arg2    U
		arg3    V
		arg4    W
		ok      bool
		err     error
	)
	if err = json.Unmarshal(params, &arg); err != nil {
		return nil, err
	}
	if arg1raw, ok = arg[b.arg1Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg1Name)
	} else if err = json.Unmarshal(arg1raw, &arg1); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg1Name, err)
	}
	if arg2raw, ok = arg[b.arg2Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg2Name)
	} else if err = json.Unmarshal(arg2raw, &arg2); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg2Name, err)
	}
	if arg3raw, ok = arg[b.arg3Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg3Name)
	} else if err = json.Unmarshal(arg3raw, &arg3); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg3Name, err)
	}
	if arg4raw, ok = arg[b.arg4Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg4Name)
	} else if err = json.Unmarshal(arg4raw, &arg4); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg4Name, err)
	}
	return b.toolFn(ctx, arg1, arg2, arg3, arg4)
}

// Name implements Tool.
func (b *toolType4[T, U, V, W]) Name() string {
	return b.name
}

// Parameters implements Tool.
func (b *toolType4[T, U, V, W]) Parameters() json.RawMessage {
	return b.inputSchema
}

type toolType5[T, U, V, W, X any] struct {
	toolFn                                           func(ctx context.Context, arg1 T, arg2 U, arg3 V, arg4 W, arg5 X) (json.RawMessage, error)
	name                                             string
	arg1Name, arg2Name, arg3Name, arg4Name, arg5Name string
	description                                      string
	inputSchema                                      json.RawMessage
}

var _ Tool = &toolType5[any, any, any, any, any]{}

func NewTool5[T, U, V, W, X any](name, description, arg1Name, arg2Name, arg3Name, arg4Name, arg5Name string, inputSchema json.RawMessage, toolFn func(ctx context.Context, arg1 T, arg2 U, arg3 V, arg4 W, arg5 X) (json.RawMessage, error)) *toolType5[T, U, V, W, X] {
	return &toolType5[T, U, V, W, X]{
		name:        name,
		arg1Name:    arg1Name,
		arg2Name:    arg2Name,
		arg3Name:    arg3Name,
		arg4Name:    arg4Name,
		arg5Name:    arg5Name,
		description: description,
		inputSchema: inputSchema,
		toolFn:      toolFn,
	}
}

// Description implements Tool.
func (b *toolType5[T, U, V, W, X]) Description() string {
	return b.description
}

// Execute implements Tool.
func (b *toolType5[T, U, V, W, X]) Execute(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var (
		arg     map[string]json.RawMessage
		arg1raw json.RawMessage
		arg2raw json.RawMessage
		arg3raw json.RawMessage
		arg4raw json.RawMessage
		arg5raw json.RawMessage
		arg1    T
		arg2    U
		arg3    V
		arg4    W
		arg5    X
		ok      bool
		err     error
	)
	if err = json.Unmarshal(params, &arg); err != nil {
		return nil, err
	}
	if arg1raw, ok = arg[b.arg1Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg1Name)
	} else if err = json.Unmarshal(arg1raw, &arg1); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg1Name, err)
	}
	if arg2raw, ok = arg[b.arg2Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg2Name)
	} else if err = json.Unmarshal(arg2raw, &arg2); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg2Name, err)
	}
	if arg3raw, ok = arg[b.arg3Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg3Name)
	} else if err = json.Unmarshal(arg3raw, &arg3); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg3Name, err)
	}
	if arg4raw, ok = arg[b.arg4Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg4Name)
	} else if err = json.Unmarshal(arg4raw, &arg4); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg4Name, err)
	}
	if arg5raw, ok = arg[b.arg5Name]; !ok {
		return nil, fmt.Errorf("argument %s not found", b.arg5Name)
	} else if err = json.Unmarshal(arg5raw, &arg5); err != nil {
		return nil, fmt.Errorf("unmarshaling argument %s: %w", b.arg5Name, err)
	}
	return b.toolFn(ctx, arg1, arg2, arg3, arg4, arg5)
}

// Name implements Tool.
func (b *toolType5[T, U, V, W, X]) Name() string {
	return b.name
}

// Parameters implements Tool.
func (b *toolType5[T, U, V, W, X]) Parameters() json.RawMessage {
	return b.inputSchema
}
