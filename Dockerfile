# build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG BUILD_TAGS=""
RUN go build -tags "$BUILD_TAGS" -o server cmd/server/main.go

# runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config.yaml .
CMD ["./server"]
