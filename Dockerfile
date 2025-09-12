FROM registry.js.design/library/golang:1.23.3-otel-amd64 AS builder

ENV GOPROXY=https://goproxy.cn,direct

RUN mkdir -p /app

WORKDIR /app

COPY . .

RUN go mod download && \
    go mod tidy && \
    otel go build -o tracer . || { cat .otel-build/preprocess/debug.log; exit 1; } && \
    chmod 777 tracer

FROM registry.js.design/base/ubuntu:22.04-amd64

WORKDIR /app

COPY --from=builder /app/tracer /app/tracer

WORKDIR /app

CMD ["/app/tracer"]
