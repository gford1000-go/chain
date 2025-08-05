package chain

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"time"
)

// ErrContextDone is raised when context is done whilst the chain is still being completed
// Context status is checked prior to invoking each func in the chain, and should also be
// checked during long running funcs.
var ErrContextDone = errors.New("context is Done()")

// Func is the type of func that can be passed to Chain.Then
type Func func(context.Context, ...any) ([]any, error)

// FinalFunc is the type of func that must be passed to Chain.Finally to generate the output
type FinalFunc[T any] func(context.Context, ...any) (T, error)

// Retry allows options to be set when retries are required
type Retry struct {
	// NumRetries specifies the max number to attempt.  Min = 0 (no retry); Max = 8.  Default = 0
	NumRetries int
	// BaseWait specifies the base sleep duration, which will be exponentially increased.
	// Default = 10ms.  Max = 1s
	BaseWait time.Duration
	// Forward specifies the errors which if encountered, are to be forwarded with no retry attempt
	// so that they are observable and acted upon.  The existence test uses via errors.Is().
	// If nil or empty slice, then all errors are silently absorbed and the function retried
	Forward []error
}

func (r Retry) ensureValid() Retry {
	out := r
	if out.NumRetries < 0 {
		out.NumRetries = 0
	}
	if out.NumRetries > 8 {
		out.NumRetries = 8
	}

	if out.BaseWait <= 0 {
		out.BaseWait = 10 * time.Millisecond
	}
	if out.BaseWait > time.Second {
		out.BaseWait = time.Second
	}

	out.Forward = []error{}
	if r.Forward != nil {
		out.Forward = append(out.Forward, r.Forward...)
	}

	return out
}

// Chain holds variadic args and tracks any error in the pipeline
type Chain[T any] struct {
	ctx   context.Context
	t     T
	retry Retry
	args  []any
	err   error
}

// New starts a new pipeline with initial input values
func New[T any](ctx context.Context, args ...any) Chain[T] {
	return NewWithRetries[T](ctx, Retry{}, args...)
}

// NewWithRetries supports transitory failures via the configured retry options
func NewWithRetries[T any](ctx context.Context, retry Retry, args ...any) Chain[T] {
	return Chain[T]{ctx: ctx, args: args, retry: retry.ensureValid()}
}

// Process is a single line equivalent for a chain call
func Process[T any](ctx context.Context, fs []Func, fn FinalFunc[T], args ...any) (T, error) {
	return ProcessWithRetries(ctx, fs, fn, Retry{}, args...)
}

// ProcessWithRetries is a single line equivalent for a chain call using retries
func ProcessWithRetries[T any](ctx context.Context, fs []Func, fn FinalFunc[T], retry Retry, args ...any) (T, error) {

	var c = NewWithRetries[T](ctx, retry, args...)

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
		result, err := c.thenWrap(f)
		if err != nil {
			funcName := runtimeFuncName(f)
			return Chain[T]{err: fmt.Errorf("error in %s: %w", funcName, err)}
		}

		return Chain[T]{args: result, t: c.t, ctx: c.ctx, retry: c.retry}
	}
}

// ErrUnhandledPanic raised if funcs panic when invoked by Then or Finally
var ErrUnhandledPanic = errors.New("unhandled panic")

// ErrExceededRetries raised if the func repeatedly returns error.  Note that if the
// func panics then retries are not attempted
var ErrExceededRetries = errors.New("exceeded retry count")

func (c Chain[T]) thenWrap(f Func) (result []any, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("%v: %w", r, ErrUnhandledPanic)
		}
	}()

	attempt := 0
	for range 1 + c.retry.NumRetries {
		if result, err := f(c.ctx, c.args...); err == nil {
			return result, err
		} else {
			if c.retry.NumRetries == 0 {
				return nil, err
			}
			for _, e := range c.retry.Forward {
				if errors.Is(err, e) {
					return nil, err
				}
			}
		}

		c.sleep(attempt)
		attempt++
	}

	return nil, ErrExceededRetries
}

func (c Chain[T]) sleep(attempt int) {
	backoff := c.retry.BaseWait * (1 << attempt) // 2^attempt

	jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
	sleep := backoff + jitter

	<-time.After(sleep)
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

		result, err := c.finallyWrap(f)
		if err != nil {
			funcName := runtimeFuncName(f)
			return c.t, fmt.Errorf("error in %s: %w", funcName, err)
		}

		return result, nil
	}
}

func (c Chain[T]) finallyWrap(f FinalFunc[T]) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			var zero T
			result = zero
			err = fmt.Errorf("%v: %w", r, ErrUnhandledPanic)
		}
	}()

	attempt := 0
	for range 1 + c.retry.NumRetries {
		if result, err := f(c.ctx, c.args...); err == nil {
			return result, err
		} else {
			if c.retry.NumRetries == 0 {
				return c.t, err
			}
			for _, e := range c.retry.Forward {
				if errors.Is(err, e) {
					return c.t, err
				}
			}
		}

		c.sleep(attempt)
		attempt++
	}

	return c.t, ErrExceededRetries
}

// Helper to get function name for debug/error reporting
func runtimeFuncName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
