# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy the source code
COPY tictactoe.go .

# Build the application
RUN go build -o tictactoe tictactoe.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/tictactoe .

# Run the game
CMD ["./tictactoe"]
