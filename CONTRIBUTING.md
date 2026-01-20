# Contributing

Want to contribute to the Kubernetes extension? We recommend opening an issue first to discuss your proposed changes. Once aligned, refer to the guide below for development workflow.

## Project Structure

```
cmd/main.go              # Entry point
pkg/extension/
  extension.go           # Extension struct, New(), Run()
  client.go              # ResourceClient interface and adapter
  resource.go            # Resource reference parsing helpers
  operations.go          # Operation registration
  create.go              # Create handler
  wait.go                # Wait handler
  delete.go              # Delete handler
  *_test.go              # Unit tests
```

## Adding a New Operation

1. Create `pkg/extension/<operation>.go` with your handler:

```go
func (e *Extension) handleMyOp(ctx context.Context, req *sdk.OperationRequest) (*sdk.OperationResult, error) {
    if e.client == nil {
        return sdk.Failure(fmt.Errorf("kubernetes client not initialized")), nil
    }

    args, ok := req.Args.(map[string]any)
    if !ok {
        return sdk.Failure(fmt.Errorf("args must be an object")), nil
    }

    // Your logic here

    e.LogInfo(ctx, "Operation completed", map[string]any{"key": "value"})
    return sdk.Success("Done"), nil
}
```

2. Register the operation in `operations.go`:

```go
e.AddOperation(
    sdk.NewOperation("myop",
        sdk.WithDescription("Description of your operation"),
        sdk.WithParams(jsonschema.Schema{
            Type: "object",
            Properties: map[string]*jsonschema.Schema{
                "field": {Type: "string", Description: "Field description"},
            },
            Required: []string{"field"},
        }),
    ),
    e.handleMyOp,
)
```

3. Add tests in `pkg/extension/<operation>_test.go` using table-driven tests.

## Testing

Run all tests:
```bash
go test ./...
```

Run with verbose output:
```bash
go test ./... -v
```

## Code Style

- Use `e.LogInfo()` and `e.LogError()` for logging
- Return `sdk.Failure(err)` for errors, not Go errors
- Use `parseResourceRef()` for standard resource arguments
- Keep handlers focused on a single responsibility
