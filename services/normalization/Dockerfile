# Multi-stage build for the normalization-service.
# Build context is this service directory (services/normalization).
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static, stripped binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/normalization ./cmd/normalization

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/normalization /app/normalization

# Documented service port (NORMALIZATION_PORT). All real config is supplied at
# runtime via environment variables — nothing is baked into the image. Mount the
# affiliate config and point NORMALIZATION_AFFILIATE_CONFIG at it.
EXPOSE 8084
USER nonroot:nonroot
ENTRYPOINT ["/app/normalization"]
