# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25.1

FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/user-http ./cmd/user-http
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/user-grpc ./cmd/user-grpc

FROM gcr.io/distroless/base-debian12:nonroot AS runtime

WORKDIR /app
COPY --from=builder /out/user-http /usr/local/bin/user-http
COPY --from=builder /out/user-grpc /usr/local/bin/user-grpc

EXPOSE 8080 9090

ENTRYPOINT ["/usr/local/bin/user-http"]
