# Build stage
FROM golang:1.20-bookworm AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o ratelimiter ./cmd/ratelimiter/main.go

# Final stage - minimal image
FROM alpine:latest

# Install dumb-init in Alpine for signal handling
RUN apk --no-cache add dumb-init

# Copy the built binary from the build stage
COPY --from=build /build/ratelimiter /app/ratelimiter

# Use dumb-init to handle signals properly
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Run the compiled Go binary directly
CMD ["/app/ratelimiter"]

# Expose the port if your application listens on one
EXPOSE 8080
