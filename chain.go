package chain

import (
	"fmt"
	"reflect"
	"runtime"
)

// Func is the type of func that can be passed to Chain.Then
type Func func(...any) ([]any, error)

// FinalFunc is the type of func that must be passed to Chain.Finally to generate the output
type FinalFunc[T any] func(...any) (T, error)

// Chain holds variadic args and tracks any error in the pipeline
type Chain[T any] struct {
	t    T
	args []any
	err  error
}

// New starts a new pipeline with initial input values
func New[T any](args ...any) Chain[T] {
	return Chain[T]{args: args}
}

// Then adds a transformation step: func(...any) ([]any, error)
func (c Chain[T]) Then(f Func) Chain[T] {
	if c.err != nil {
		return c
	}

	result, err := f(c.args...)
	if err != nil {
		funcName := runtimeFuncName(f)
		return Chain[T]{args: c.args, err: fmt.Errorf("error in %s: %w", funcName, err)}
	}

	return Chain[T]{args: result, t: c.t}
}

// Finally is a generic method on Chain that ends the pipeline
func (c Chain[T]) Finally(f FinalFunc[T]) (T, error) {

	if c.err != nil {
		return c.t, c.err
	}

	result, err := f(c.args...)
	if err != nil {
		funcName := runtimeFuncName(f)
		return c.t, fmt.Errorf("error in %s: %w", funcName, err)
	}

	return result, nil
}

// Helper to get function name for debug/error reporting
func runtimeFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
