# Staticlint multichecker

Custom multichecker that bundles go/analysis checks (printf, shadow, structtag), staticcheck (all SA* and S1000), bodyclose, sqlrows, and the project-specific noexit rule (forbids os.Exit in main packages).

## Build
- `go build -o ./cmd/staticlint/multichecker ./cmd/staticlint`

## Run
- From repo root after build: `./cmd/staticlint/multichecker ./...`
- Or run without building: `go run ./cmd/staticlint ./...`

## Notes
- Exits non-zero if any analyzer reports issues.
- Targets all Go packages matched by the provided patterns (use `./...` to scan the whole module).
