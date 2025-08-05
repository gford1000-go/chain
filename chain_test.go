package chain

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func ExampleNew() {

	f1 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		return []any{x + 1}, nil
	}

	f2 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		return []any{x * 2}, nil
	}

	f3 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		if x < 0 {
			return nil, errors.New("x became negative")
		}
		return []any{x - 3}, nil
	}

	f4 := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x * x, nil
	}

	input := 5

	result, _ := New[int](context.Background(), input).
		Then(f1).
		Then(f2).
		Then(f3).
		Finally(f4)

	fmt.Println("Result:", result)
	// Output: Result: 81
}

func ExampleNew_single() {

	finalOnly := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x * x, nil
	}

	result, _ := New[int](context.Background(), 5).Finally(finalOnly)

	fmt.Println("Result:", result)
	// Output: Result: 25
}

func ExampleNew_noop() {

	noOp := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x, nil
	}

	result, _ := New[int](context.Background(), 5).Finally(noOp)

	fmt.Println("Result:", result)
	// Output: Result: 5
}

func ExampleNew_failure() {

	f1 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		return []any{x + 1}, nil
	}

	f2 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		return []any{x * 2}, nil
	}

	f3 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		if x < 0 {
			return nil, errors.New("x became negative")
		}
		return []any{x - 3}, nil
	}

	f4 := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x * x, nil
	}

	input := -5

	result, err := New[int](context.Background(), input).
		Then(f1).
		Then(f2).
		Then(f3).
		Finally(f4)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Result:", result)
	// Output: error in github.com/gford1000-go/chain.ExampleNew_failure.func3: x became negative
}

func TestNew(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	f1 := func(ctx context.Context, args ...any) ([]any, error) {
		<-time.After(10 * time.Millisecond)
		return args, nil
	}

	f2 := func(ctx context.Context, args ...any) (int, error) {
		return args[0].(int), nil
	}

	go func() {
		<-time.After(5 * time.Millisecond)
		cancel()
	}()

	_, err := New[int](ctx, 5).
		Then(f1).
		Finally(f2)

	if err == nil {
		t.Fatal("expected error, not nil")
	}

	if !errors.Is(err, ErrContextDone) {
		t.Fatalf("expected context done error, got: %v", err)
	}
}

func TestProcess(t *testing.T) {

	f1 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		return []any{x + 1}, nil
	}

	f2 := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x + 2, nil
	}

	var i int = 5

	result, err := Process(context.Background(),
		[]Func{f1},
		f2,
		i)

	if err != nil {
		t.Fatalf("unexpected error, got: %v", err)
	}

	if result != 8 {
		t.Fatalf("unexpected result.  wanted: 3, got: %v", result)
	}
}

func TestProcess_1(t *testing.T) {

	f2 := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x + 2, nil
	}

	var i int = 5

	_, err := Process(context.Background(),
		[]Func{nil},
		f2,
		i)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrNilThenFunc) {
		t.Fatalf("expected NilThen error, got: %v", err)
	}
}

func TestProcess_2(t *testing.T) {

	f1 := func(ctx context.Context, args ...any) ([]any, error) {
		x := args[0].(int)
		return []any{x + 1}, nil
	}

	var i int = 5

	_, err := Process[int](context.Background(),
		[]Func{f1},
		nil,
		i)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrNilFinalFunc) {
		t.Fatalf("expected NilFinally error, got: %v", err)
	}
}

func TestProcess_3(t *testing.T) {

	f2 := func(ctx context.Context, args ...any) (int, error) {
		x := args[0].(int)
		return x + 2, nil
	}

	var i int = 5

	result, err := Process(context.Background(),
		nil,
		f2,
		i)

	if err != nil {
		t.Fatalf("unexpected error, got: %v", err)
	}

	if result != 7 {
		t.Fatalf("unexpected result.  wanted: 3, got: %v", result)
	}
}

func TestProcess_4(t *testing.T) {

	var i int = 5

	_, err := Process[int](context.Background(),
		nil,
		nil,
		i)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrNilFinalFunc) {
		t.Fatalf("expected NilFinally error, got: %v", err)
	}
}
