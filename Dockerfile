# ==============================================================================
# Stage 1: Build
# ==============================================================================
FROM golang:1.26 AS builder

WORKDIR /workspace

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY api/ api/

# Build the operator binary
# CGO_ENABLED=0 produces a static binary that runs in scratch/distroless
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -ldflags="-w -s" \
    -o manager \
    ./cmd/operator

# ==============================================================================
# Stage 2: Runtime
# ==============================================================================
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /workspace/manager .

# Run as non-root user (best practice for K8s)
USER 65532:65532

ENTRYPOINT ["/manager"]