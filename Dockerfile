# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go binary
RUN go build -o main .

# Final stage - use a smaller image for production
FROM alpine:latest
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/resources /resources  
# Copy config files

# Set executable permission
RUN chmod +x ./main

COPY news_data.json /app/news_data.json

# Run the binary
CMD ["./main"]
