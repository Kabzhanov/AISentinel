# Dockerfile for AISentinel MCP server and sidecar.
# Multi-stage build: compiles both binaries from source, ships a slim runtime image
# with both binaries on PATH. Use as:
#   docker run -i ghcr.io/kabzhanov/aisentinel serve --policy /policies/default.yaml
#   docker run -i ghcr.io/kabzhanov/aisentinel-sidecar --target /usr/local/bin/aisentinel serve
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

# Slim runtime image.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/aisentinel /usr/local/bin/aisentinel
COPY --from=build /out/aisentinel-sidecar /usr/local/bin/aisentinel-sidecar
COPY policies /policies
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/aisentinel"]
CMD ["--help"]%

# Stage 2 — build sidecar as a separate image so users can pick the right tag.
FROM scratch AS sidecar
COPY --from=build /out/aisentinel-sidecar /aisentinel-sidecar
ENTRYPOINT ["/aisentinel-sidecar"]
