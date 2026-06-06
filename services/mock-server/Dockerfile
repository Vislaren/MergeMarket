FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/mock-server ./cmd/mock-server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/mock-server /app/mock-server
EXPOSE 8089
USER nonroot:nonroot
ENTRYPOINT ["/app/mock-server"]
