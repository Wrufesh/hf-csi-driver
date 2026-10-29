# Build CSI driver + sidecar mounter (Go)
FROM golang:1.25-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 go build -o /bin/hf-csi-driver ./cmd/hf-csi-driver/

# Runtime: needs /bin/mount for FUSE mount via mount-utils
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends mount ca-certificates && rm -rf /var/lib/apt/lists/*
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
RUN printf '#!/bin/sh\nupdate-ca-certificates\nif [ "$#" -eq 0 ] || [ "${1#-}" != "$1" ]; then\n    exec /bin/hf-csi-driver "$@"\nfi\nexec "$@"\n' > /usr/local/bin/docker-entrypoint.sh && \
    chmod +x /usr/local/bin/docker-entrypoint.sh
COPY --from=builder /bin/hf-csi-driver /bin/
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["/bin/hf-csi-driver"]