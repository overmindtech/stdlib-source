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

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -trimpath -o source main.go

FROM alpine:3.20
WORKDIR /
COPY --from=builder /workspace/source .
# create our termination log folder
RUN mkdir -p /log && chown -R nobody:nobody /log
USER 65534:65534

ENTRYPOINT ["/source"]
