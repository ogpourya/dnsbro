FROM golang:1.21-bullseye AS builder
WORKDIR /src
COPY . .
RUN go build -o /usr/local/bin/dnsbro ./cmd/dnsbro

FROM debian:12-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates dnsutils && rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/local/bin/dnsbro /usr/local/bin/dnsbro
COPY configs/config.yaml /etc/dnsbro/config.yaml
EXPOSE 53/udp 53/tcp
CMD ["dnsbro", "serve", "--config", "/etc/dnsbro/config.yaml", "--no-tui"]
