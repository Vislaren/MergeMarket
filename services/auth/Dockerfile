FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth-service ./cmd/auth

FROM alpine:3.20
RUN apk add --no-cache openssl && adduser -D -H appuser && mkdir -p /certs && chown appuser:appuser /certs
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
USER appuser
COPY --from=build /out/auth-service /usr/local/bin/auth-service
EXPOSE 8091
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
