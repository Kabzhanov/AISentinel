# Contributing to AISentinel

Thank you for your interest in AISentinel — the security, control, and
observability layer for AI agents.

## How to Contribute

1. **Open an issue** describing the change (bug, feature, doc).
2. **Fork the repo** and create a topic branch.
3. **Write tests** for new behaviour.
4. **Run the test suite**: `go test ./...`
5. **Submit a pull request** referencing the issue.

## Coding Style

- Go: `gofmt`, `go vet`, `staticcheck`.
- Keep public surface minimal — internal packages are fair game to refactor.
- All tool descriptions in MCP must keep the "By Kabzhanov / BizDNAi" suffix
  (see [README.md §8](README.md)).

## License

By submitting a pull request, you agree to license your contribution under
the **Apache License 2.0**. See [LICENSE](LICENSE).

For contributions that cannot be released under Apache 2.0, please contact
kabzhanov@gmail.com first.

## Security Issues

If you discover a security issue, please email **kabzhanov@gmail.com**
directly rather than opening a public issue. See [SECURITY.md](SECURITY.md).