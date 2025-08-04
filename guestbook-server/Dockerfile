# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

## Copy source code
COPY . .

## Download dependencies
RUN go mod download

RUN ls -alh

## Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o guestbook-server ./cmd/guestbook-server

# Final stage
FROM alpine:latest

## Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

## Copy the binary from builder stage
COPY --from=builder /app/guestbook-server .

## Expose port
EXPOSE 8080

## Run the binary
CMD ["./guestbook-server"]
