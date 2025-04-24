# Use the official Golang image as the build environment
FROM golang:1.22.4-alpine AS builder

# Install git (required for Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /source

# Copy the source code
COPY . .

# Download dependencies
RUN go mod download

# Build the Go CLI binary
RUN GOOS=linux GOARCH=amd64 go build -o /source/bin/mdx .

# Final minimal image
FROM alpine:3.20

RUN mkdir /app

COPY --from=builder /source/bin/mdx /app

RUN chmod +x /app/mdx

ENTRYPOINT ["./app/mdx"]