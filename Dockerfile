FROM golang:1.25 AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o reader ./cmd/reader

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /usr/src/app/reader .

USER nonroot:nonroot

CMD ["./reader"]
