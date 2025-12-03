# Build stage
FROM registry.access.redhat.com/ubi9/go-toolset:1.24 AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies
RUN go mod download

# Copy source code
COPY main.go main.go

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o controller main.go

# Final stage - using Red Hat UBI micro for minimal image size
FROM registry.access.redhat.com/ubi9/ubi-micro:latest

WORKDIR /

# Copy the binary from builder
COPY --from=builder /workspace/controller .

# Use non-root user
USER 65532:65532

ENTRYPOINT ["/controller"]
