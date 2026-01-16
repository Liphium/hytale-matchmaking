# Use official Golang image as base image
FROM golang:1.25.5-bookworm AS builder
WORKDIR /app

# Copy source code
COPY . .

# Build the Go app
RUN go mod download
RUN go build -tags release -o start .

# Make the final container based on a small image
FROM debian:bookworm
WORKDIR /app

# Make sure certificates work properly
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates
RUN apt-get install -y curl
RUN apt-get install -y wget

# Copy the current executable over to the container from the builder
COPY --from=builder /app/start .

# Create a volume for persistent token storage
VOLUME ["/app/tokens.json"]

# Run the app together with the ports
EXPOSE 80
CMD ["./start"]
