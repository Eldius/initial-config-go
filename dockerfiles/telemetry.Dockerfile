FROM golang:1.25.5-alpine3.23 AS builder
RUN addgroup -S nonroot && \
    adduser -S nonroot -G nonroot

RUN apk update && apk add --no-cache git openssh make

WORKDIR /app

ADD .. /app
RUN go mod tidy
RUN echo "$(git rev-parse --short HEAD)"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -a \
      -v \
      -ldflags \
      "-X 'main.AppVersion=$(git rev-parse --short HEAD)' -extldflags '-static'" \
      -o /app/otel \
        /app/examples/telemetry/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
USER nonroot
WORKDIR /app
COPY --from=builder /app/examples/telemetry/config.yaml config.yaml
COPY --from=builder /app/otel otel
ENTRYPOINT ["/app/otel"]
