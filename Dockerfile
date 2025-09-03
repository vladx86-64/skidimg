# Dockerfile for skidimg image hosting with bimg + libvips
FROM golang:1.24.2

# Install libvips and dependencies
RUN apt-get update && \
    apt-get install -y libvips-dev pkg-config && \
    apt-get clean

# Set working directory
WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go app
RUN go build -o skidimg ./cmd/main.go

# Expose the server port
EXPOSE 8080

# Run the app
CMD ["./skidimg"]
