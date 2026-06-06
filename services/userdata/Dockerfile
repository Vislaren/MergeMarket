# Multi-stage build for the userdata-service.
# Build context is this service directory (services/userdata).
FROM golang:1.22-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static, stripped binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/userdata ./cmd/userdata

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/userdata /app/userdata

# Documented service port (USERDATA_PORT). All real config is supplied at runtime
# via environment variables — nothing is baked into the image.
EXPOSE 8090
USER nonroot:nonroot
ENTRYPOINT ["/app/userdata"]
