# Dockerfile for AISentinel MCP server and sidecar.
# Multi-stage build produces two images:
#   * default target    -> AISentinel MCP server (distroless, both binaries + policies)
#   * --target sidecar   -> AISentinel policy-proxy sidecar (scratch)
# Use as:
#   docker run -i ghcr.io/kabzhanov/aisentinel serve                                  # MCP server
#   docker run -i ghcr.io/kabzhanov/aisentinel-sidecar --policy /policies/strict.yaml your-mcp-server
#
# For drop-in MCP client integration, see docker-compose.yml.

FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache modules first.
COPY go.mod go.sum ./
RUN go mod download

# Build both binaries.
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/aisentinel ./cmd/aisentinel
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/aisentinel-sidecar ./cmd/aisentinel-sidecar

# Sidecar image (policy proxy) — build with `--target sidecar`.
FROM scratch AS sidecar
COPY --from=build /out/aisentinel-sidecar /aisentinel-sidecar
COPY policies /policies
ENTRYPOINT ["/aisentinel-sidecar"]

# AISentinel MCP server — DEFAULT target (must remain the LAST stage).
FROM gcr.io/distroless/static-debian12:nonroot AS server
WORKDIR /
COPY --from=build /out/aisentinel /usr/local/bin/aisentinel
COPY --from=build /out/aisentinel-sidecar /usr/local/bin/aisentinel-sidecar
COPY policies /policies
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/aisentinel"]
CMD ["serve"]
