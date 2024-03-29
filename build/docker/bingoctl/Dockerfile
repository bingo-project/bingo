FROM golang:latest as builder
WORKDIR /workspace

ARG OS
ARG ARCH

# Copy the Go Modules manifests
COPY go.mod .
COPY go.sum .

RUN go mod download

# Copy the sources
COPY api/ api/
COPY cmd/bingoctl cmd/bingoctl
COPY pkg/ pkg/
COPY internal/ internal/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${OS:-linux} GOARCH=${ARCH:-amd64} go build -a ./cmd/bingoctl

FROM alpine:latest
LABEL maintainer="<brooksyang@outlook.com>"

WORKDIR /opt/bingo

COPY --from=builder /workspace/bingoctl bin/

ENTRYPOINT ["/opt/bingo/bin/bingoctl"]
