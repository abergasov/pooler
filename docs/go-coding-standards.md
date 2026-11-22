# Go coding standards

Adapted from `turbonet-poc/AGENTS.md` and `.golangci.yaml`.
These requirements are mandatory for Go code changes in Pooler. They do not
specify smart contract behavior or require adding Go services before they are
needed.

## Formatting and linting

- Use the repository-root `.golangci.yaml` for formatting and linting.
- Format code with `gofumpt` and organize imports with `gci` through
  `golangci-lint fmt`. Keep standard-library, third-party, and local `pooler`
  imports in separate groups, in that order.
- Fix enabled linter findings. Explain any necessary `nolint` directive.
- Do not add `github.com/sirupsen/logrus` or `github.com/pkg/errors`.

## Production code

- Keep service code under `internal/...`; keep entry points in `cmd/` or `main.go`.
- Wrap errors with context using `%w`, for example:
  `fmt.Errorf("fetch job %d: %w", id, err)`. Do not use `github.com/pkg/errors`.
- Use `github.com/getoptimum/optimum-common/pkg/logger` for application logging.
  Log only through `logger.AppLogger` (`Info`, `Warn`, `Error`, `Fatal`, `With`)
  and the relevant `logger.With*` helpers, such as `WithPackage`,
  `WithFunctionName`, `WithHeight`, and `WithDuration`.
- Pass `ctx context.Context` as the first parameter and propagate it through the
  call chain. Do not create `context.Background()` inside a call chain.
- Inject dependencies through constructors, store them in container structs,
  and assemble containers at startup. Do not use package-level singletons.
- Constructors return concrete objects, not interfaces. Define an interface only
  at the consuming boundary when polymorphism is actually required.
- An exported function must not return an unexported type. Export the concrete
  return type as part of the package API.
- Keep one function per operation. Do not add a private implementation and an
  exported wrapper with the same signature when the exported function can
  contain the implementation directly.
- Keep one public lifecycle path per component. A `Start` or `Run` method must
  own its complete listen/serve/shutdown flow; do not add alternate public
  lifecycle methods only to inject test internals such as `net.Listener`.
- Do not split short lifecycle code into helpers unless the helper is reused or
  isolates a distinct operation with its own error context.
- Propagate unrecoverable component errors to the application entry point. The
  top-level orchestrator must cancel sibling components, and `main` must report
  the final error with `logger.Fatal`. Do not call `logger.Fatal` below the
  application boundary because it terminates the process and bypasses deferred
  cleanup.

## Tests

- Use only `require` from `testify`; do not use `assert`.
- Structure tests with `// given`, `// when`, and `// then`. Combine stages only
  when one is genuinely empty, for example `// when, then`.
- Put tests in an external `package foo_test`.
- Do not add `export_internal_test.go` or bridge files to expose internals to
  tests. A narrowly scoped, documented public method is allowed only when it
  executes real production behavior needed by an external-package test.
- Use real Docker-backed dependencies. Use mocks only when the task explicitly
  requires them.

## Verification

After Go code changes, run formatting, focused regression tests, and relevant
static checks before completing the work:

```sh
golangci-lint fmt
golangci-lint run ./...
go test ./...
```

Use `go test -race ./...` for changes involving concurrency. Integration tests
must have their real dependencies available; report unavailable checks and
known exclusions explicitly.

These commands apply once the repository contains Go packages. The source
repository's `scripts/check_structure.sh` is not included: it hard-codes its
component names and infrastructure assumptions, and references a missing
`doc/structure.md`.
