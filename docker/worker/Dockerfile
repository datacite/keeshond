# Builder
FROM golang:1.22.4 AS builder

# Echo pwd
RUN echo ${PWD}

COPY ${PWD} /app
WORKDIR /app

# Build the go environment
RUN CGO_ENABLED=0 go build -ldflags '-s -w -extldflags "-static"' -o /app/workerbin ./cmd/worker

# Production image
FROM alpine:latest

# Install CA certs
RUN apk --no-cache add ca-certificates

# Add a new non root user to run the app under
RUN adduser --home "/app" --disabled-password app --gecos "app,-,-,-"
USER app

# Copy the go binary from the builder to the container
COPY --chown=app:app --from=builder /app/workerbin /home/app/workerbin

# Set the workdir to app dir
WORKDIR /home/app/

# Default execution
CMD ["./workerbin"]