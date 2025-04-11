# Use an official Go runtime as a parent image
FROM golang:1.24 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Use a smaller image for the final stage
# FROM alpine:latest

# Set the working directory inside the container
# WORKDIR /app

# Copy the binary from the builder stage
# COPY --from=builder /app/main .

# Copy the database file if exist
#COPY --from=builder /app/alpha-enigma.db .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
