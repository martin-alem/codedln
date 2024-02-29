# Stage 1: Build the application
# Use the official Golang image for the building stage
FROM golang:1.21 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy all file to current directory
COPY . .

# Download module dependencies
RUN go mod download


# Build the application
# -o codedln sets the output name of the binary
RUN CGO_ENABLED=0 GOOS=linux go build -v -o codedln ./cmd/main/main.go

# Stage 2: Build a small image for runtime
# Use a small base image
FROM alpine:latest

# Set the working directory in the new container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/codedln .

# Expose port (adjust if different)
EXPOSE 8080

# Run the binary
CMD ["./codedln"]
