# syntax = docker/dockerfile:1-experimental

FROM golang:1.21 as sleeper-builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -ldflags "-w" -o sleeper ./cmd/sleeper.go

FROM scratch
COPY --from=sleeper-builder /etc/passwd /etc/passwd
COPY --from=sleeper-builder /app/sleeper /

ENTRYPOINT ["/sleeper"]
