# Builder
FROM golang:1.17 AS builder

COPY ${PWD} /app
WORKDIR /app

# Build the go environment
RUN CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o /app/webbin ./cmd/web

# Production image
FROM alpine:latest

# Install CA certs
RUN apk --no-cache add ca-certificates

# Add a new non root user to run the app under
RUN adduser --home "/app" --disabled-password app --gecos "app,-,-,-"
USER app

# Copy the go binary from the builder to the container
COPY --chown=app:app --from=builder /app/webbin /home/app/webbin

# Set the workdir to app dir
WORKDIR /home/app/

# Default execution
CMD ["./webbin"]

# Expose the webserver
EXPOSE 8081