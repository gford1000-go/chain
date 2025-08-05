package chain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// ErrContextDone is raised when context is done whilst the chain is still being completed
// Context status is checked prior to invoking each func in the chain, and should also be
// checked during long running funcs.
var ErrContextDone = errors.New("context is Done()")

// Func is the type of func that can be passed to Chain.Then
type Func func(context.Context, ...any) ([]any, error)

// FinalFunc is the type of func that must be passed to Chain.Finally to generate the output
type FinalFunc[T any] func(context.Context, ...any) (T, error)

// Chain holds variadic args and tracks any error in the pipeline
type Chain[T any] struct {
	ctx  context.Context
	t    T
	args []any
	err  error
}

// New starts a new pipeline with initial input values
func New[T any](ctx context.Context, args ...any) Chain[T] {
	return Chain[T]{ctx: ctx, args: args}
}

// Process is a single line equivalent for a chain call
func Process[T any](ctx context.Context, fs []Func, fn FinalFunc[T], args ...any) (T, error) {

	c := New[T](ctx, args)

	for _, f := range fs {
		c = c.Then(f)
	}

	return c.Finally(fn)
}

// ErrNilThenFunc is raised if a nil func is passsed to Then
var ErrNilThenFunc = errors.New("func provided to Then cannot be nil")

// Then adds a transformation step: func(...any) ([]any, error)
func (c Chain[T]) Then(f Func) Chain[T] {
	if c.err != nil {
		return c
	}
	if f == nil {
		return Chain[T]{err: ErrNilThenFunc}
	}

	select {
	case <-c.ctx.Done():
		funcName := runtimeFuncName(f)
		return Chain[T]{err: fmt.Errorf("prior to call to %s, %w", funcName, ErrContextDone)}
	default:
		result, err := f(c.ctx, c.args...)
		if err != nil {
			funcName := runtimeFuncName(f)
			return Chain[T]{err: fmt.Errorf("error in %s: %w", funcName, err)}
		}

		return Chain[T]{args: result, t: c.t, ctx: c.ctx}
	}
}

// ErrNilFinalFunc is raised if a nil func is passsed to Finally
var ErrNilFinalFunc = errors.New("func provided to Finally cannot be nil")

// Finally is a generic method on Chain that ends the pipeline
func (c Chain[T]) Finally(f FinalFunc[T]) (T, error) {
	if c.err != nil {
		return c.t, c.err
	}
	if f == nil {
		return c.t, ErrNilFinalFunc
	}

	select {
	case <-c.ctx.Done():
		funcName := runtimeFuncName(f)
		return c.t, fmt.Errorf("prior to call to %s, %w", funcName, ErrContextDone)
	default:

		result, err := f(c.ctx, c.args...)
		if err != nil {
			funcName := runtimeFuncName(f)
			return c.t, fmt.Errorf("error in %s: %w", funcName, err)
		}

		return result, nil
	}
}

// Helper to get function name for debug/error reporting
func runtimeFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
