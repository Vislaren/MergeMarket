# Multi-stage build for the proxy-validator service.
# Build context is this service directory (services/proxy-validator).
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static, stripped binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/proxy-validator ./cmd/proxy-validator

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/proxy-validator /app/proxy-validator

# Documented service port (PROXY_VALIDATOR_PORT). All real config is supplied
# at runtime via environment variables — nothing is baked into the image.
EXPOSE 8086
USER nonroot:nonroot
ENTRYPOINT ["/app/proxy-validator"]
