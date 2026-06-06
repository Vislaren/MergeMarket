# Multi-stage build for the search-service.
# Build context is this service directory (services/search).
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static, stripped binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/search ./cmd/search

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/search /app/search

# Documented service port (SEARCH_PORT). All real config is supplied at runtime
# via environment variables — nothing is baked into the image.
EXPOSE 8087
USER nonroot:nonroot
ENTRYPOINT ["/app/search"]
