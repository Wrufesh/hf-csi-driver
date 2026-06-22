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
RUN apt-get update && apt-get install -y --no-install-recommends mount && rm -rf /var/lib/apt/lists/*
COPY --from=builder /bin/hf-csi-driver /bin/

# TRUST INTERNAL ACCELERATOR S3
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY <<EOF /usr/local/share/ca-certificates/harica-geant.crt
-----BEGIN CERTIFICATE-----
MIIDtjCCAp4CCQDFm01lBlHbcDANBgkqhkiG9w0BAQsFADCBnDELMAkGA1UEBhMC
QVQxEDAOBgNVBAgMB0F1c3RyaWExEjAQBgNVBAcMCUxheGVuYnVyZzEOMAwGA1UE
CgwFSUlBU0ExDDAKBgNVBAsMA0lDVDEgMB4GA1UEAwwXY2VydGlmaWNhdGUuaWlh
c2EuYWMuYXQxJzAlBgkqhkiG9w0BCQEWGGljdC5oZWxwZGVza0BpaWFzYS5hYy5h
dDAeFw0yMzAzMjMxMzA3MDVaFw00MzEwMDQxMzA3MDVaMIGcMQswCQYDVQQGEwJB
VDEQMA4GA1UECAwHQXVzdHJpYTESMBAGA1UEBwwJTGF4ZW5idXJnMQ4wDAYDVQQK
DAVJSUFTQTEMMAoGA1UECwwDSUNUMSAwHgYDVQQDDBdjZXJ0aWZpY2F0ZS5paWFz
YS5hYy5hdDEnMCUGCSqGSIb3DQEJARYYaWN0LmhlbHBkZXNrQGlpYXNhLmFjLmF0
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA08pA8TlPOhQ1rg2zBXgy
2ZOAPSB1GKsxuLhgqRh9MxBkfBKqqwbuvt2r/DFrqOccKY2njgKdwmxweqcp2T/H
hH756LOHiEZNvv6zBodpkYMF+VxSkepVTPIvNdHCFvy12c2uM4dL7pHhOqVBf6Ly
2wfmP/fj0mwJeRLx8wDvyMUkKf3kC6UTvT5AbK0LI6jeyLxJlzF6YQqGK6L52RS1
Pbnu4gIODHJHsNshg1QmBCQYI6v1L4FXgosNbksPf05wL2SB+DI/kktLP8qSXtkx
IV7WBPsilnu8R0md2wHL+WUNTwmukB2W6KlRqoqSgZJ3nNRaqnOq8HLU3FR0Fjn9
JwIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQAexeWy9rEH3x0SLK2D8VBgggIJv3iY
ZPeMMAotF9fop/+Tf4KrTs3tbs4mwDmg9dlxMNlAYvdOyC1mfSfg5qjCF71WRxY2
a+9sIb2rvmaQ5pEuO8i7RGgTOeHj5E7f8UoCwRnC+JUw52eOTjcCfw1QxoWGieiB
whrNbNhjI0xWDNxLb2VZ0rfFtO6lEFzVQbF6GIXq4QOjxtWRV/DQKX+S4aZgmniT
0vTP1bVoS1vHkidVAFZ9v82pCGZFXpjku/gjjmO4Yc/In/WeqiyKZ2HRzIO/ZcGk
nP6/j/YnBT9ayxJE5ku2OXNh/EiuNZRytdImcik6K4TePQjhvP4gXmK5
-----END CERTIFICATE-----
EOF

RUN update-ca-certificates

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV SSL_CERT_DIR=/etc/ssl/certs
# END TRUST INTERNAL ACCELERATOR S3

ENTRYPOINT ["/bin/hf-csi-driver"]
