package chain

import (
	"context"
	"errors"
	"fmt"
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
