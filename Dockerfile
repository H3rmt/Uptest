# Use the official Go image as the base image for building the application
FROM golang:1.21.5 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and install the Go dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o .

# Use a alpine container as the base image for the final application
FROM alpine:latest

# Copy the built Go application from the builder stage
COPY --from=builder /app/Uptest /
COPY index.html /
COPY favicon.ico /
COPY style.css /
COPY info.json /

# Set the entry point for the container
ENTRYPOINT ["/Uptest"]
