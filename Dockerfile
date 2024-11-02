# Dockerfile
FROM golang:1.23.2-alpine3.20

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Run the application
CMD ["./main"]
