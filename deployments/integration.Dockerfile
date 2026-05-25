FROM golang:1.26.3-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["go", "test", "-v", "./tests/integration/..."]
