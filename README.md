[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://en.wikipedia.org/wiki/MIT_License)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/gford1000-go/chain)

# Chain

Controlled chaining of `func`s together using the Flow representation to produce a desired output.

```go
func main() {
    first := func(ctx context.Context, args ...any) ([]any, error) {
        x := args[0].(int)
        return []any{x + 1}, nil
    }

    finally := func(ctx context.Context, args ...any) (int, error) {
        x := args[0].(int)
        return x * x, nil
    }

    ctx := context.Background()
    input := 5

    result, _ := chain.New[int](ctx, input).
        Then(first).
        Finally(finally)

    fmt.Println("Result:", result)
    // Output: Result: 36
}
```

The `Process` function can be used as an alternative.

```go
func main() {
    f := func(ctx context.Context, args ...any) ([]any, error) {
        x := args[0].(int)
        return []any{x + 1}, nil
    }

    finally := func(ctx context.Context, args ...any) (int, error) {
        x := args[0].(int)
        return x * x, nil
    }

    ctx := context.Background()
    input := 5

    result, _ := chain.Process(ctx, []chain.Func{f}, finally, input)

    fmt.Println("Result:", result)
    // Output: Result: 36
}
```

The chain will check for context completion between calls to each function, exiting the chain should this occur with an appropriate error.  Long running functions should similarly check for context completion.

Should any of the functions generate an unhandled `panic`, the chain will capture the details of the panic and return as an error.

See examples for usage.
