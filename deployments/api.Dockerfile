FROM golang:1.26.3-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG LDFLAGS="-s -w"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="${LDFLAGS}" -o anti-bruteforce ./cmd/anti-bruteforce

FROM alpine:latest
RUN apk --no-cache add ca-certificates

RUN addgroup -g 10001 -S appgroup && \
    adduser -u 10001 -S appuser -G appgroup

WORKDIR /app
RUN chown -R appuser:appgroup /app

COPY --from=builder --chown=appuser:appgroup /app/anti-bruteforce .
COPY --from=builder --chown=appuser:appgroup /app/migrations ./migrations
COPY --from=builder --chown=appuser:appgroup /app/configs/config.yaml ./configs/config.yaml
USER appuser

EXPOSE 8080 50051
CMD ["./anti-bruteforce", "-config", "configs/config.yaml"]
