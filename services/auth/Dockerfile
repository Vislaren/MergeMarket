FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth-service ./cmd/auth

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/auth-service /usr/local/bin/auth-service
EXPOSE 8091
ENTRYPOINT ["/usr/local/bin/auth-service"]
