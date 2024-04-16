# Simple Golang API App

A VERY simplified calorie tracker with authentication via external GRPC (API `accounts/login` and `accounts/registrations` is just proxy for external GRPC Service)

- Go 1.22
- Layered ("onion") app architecture
- GRPC
- Taskfile
- SQLite3
- Chi router
- Mockery
- CleanEnv
- Validator

**TODO:**
- *OpenApi docs (Swagger)*
- *Full test coverage*

## How to run it in local

1. Create `local.yaml` config in `config/` dir (watch `/internal/config/config.go` and  `_example.yaml` in config dir for fields)
2. `task migrate`
3. `go run ./cmd/simple-diet-tracker/main.go --config=./config/local.yaml`
