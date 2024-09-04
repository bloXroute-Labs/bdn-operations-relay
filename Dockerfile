# Stage 1: Build stage
FROM golang:1.22.5-alpine AS build

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

CMD find .
RUN ls -la

# Build the Go application
RUN go build -o bdn-operaions-realy ./cmd/...

# Stage 2: Final stage
FROM alpine:edge

# Set the working directory
WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /app/bdn-operaions-realy .
COPY --from=build /app/config.yml .

# Set the entrypoint command
ENTRYPOINT ["/app/bdn-operaions-realy", "relay"]