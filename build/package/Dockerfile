# Build the source binary
FROM golang:1.23-alpine as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY cmd/ cmd/
COPY sources/ sources/

# Install docgen fo generating the API documentation
RUN go install github.com/overmindtech/docgen@latest

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
    go generate ./... \
    && GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -trimpath -o source main.go

FROM alpine:3.20
WORKDIR /
COPY --from=builder /workspace/source .
USER 65534:65534

ENTRYPOINT ["/source"]
