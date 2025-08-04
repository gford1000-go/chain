[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://en.wikipedia.org/wiki/MIT_License)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/gford1000-go/chain)

# Chain

Controlled chaining of `func`s together to produce a desired output.

```go
func main() {
    first := func(args ...any) ([]any, error) {
        x := args[0].(int)
        return []any{x + 1}, nil
    }

    finally := func(args ...any) (int, error) {
        x := args[0].(int)
        return x * x, nil
    }

    input := 5

    result, _ := New[int](input).
        Then(first).
        Finally(finally)

    fmt.Println("Result:", result)
    // Output: Result: 36
}
```

See examples for usage.
