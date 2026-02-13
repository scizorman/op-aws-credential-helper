# op-aws-credential-helper

A `credential_process` helper for AWS CLI. Retrieves AWS credentials from 1Password, performs MFA authentication, and returns temporary credentials.

## Architecture

### External Dependencies

- `op` (1Password CLI) — Used to retrieve credentials. Invoked as an external command.
- AWS STS — Uses Go SDK v2 (`github.com/aws/aws-sdk-go-v2`). No dependency on `aws` CLI.

The 1Password Go SDK's desktop app integration is beta and does not support WSL, so we invoke the `op` command externally.

### credential_process Constraints

`credential_process` is a protocol that returns JSON on stdout. Nothing other than credential JSON may be written to stdout. Error messages and debug logs must go to stderr. Since the caller captures stdin and stderr, interactive input requires opening `/dev/tty` directly.

### Design Principles

**Layer Separation**

Separate the presentation layer (CLI argument parsing, config loading, dependency wiring, stdout output) from domain logic. Convert config values into domain types before passing them to the domain layer.

**Interface Criteria**

Only introduce an interface when the implementation detail is outside the Resolver's concern, or when multiple implementations are realistically expected. Do not introduce interfaces based on speculation about future changes.

**Test Doubles**

Do not use mock generation tools such as mockgen. Write fakes (lightweight working implementations) by hand in `_test.go` files.

### Package Policy

Place files flat in the root package. Split into separate packages only when the code outgrows a single package. Do not use the `internal/` directory.

## Development Commands

| Target | Description |
|-----------|------|
| `make all` | Show Makefile contents (help) |
| `make clean` | Remove `result` symlink |
| `make lint` | golangci-lint |
| `make fmt` | treefmt (gofmt + nixfmt) |
| `make test` | go test |
| `make build` | nix build |
| `make ci` | lint + nix flake check |

Development environment is managed with a Nix flake. Enter the devShell with `nix develop` or direnv.
