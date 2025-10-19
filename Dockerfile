# ----------------------------
# Stage 1: Frontend build
# ----------------------------
FROM oven/bun:1 AS frontend

WORKDIR /web
COPY web/ .

# install wget and unzip
RUN apt-get update && apt-get install -y --no-install-recommends wget unzip \
 && rm -rf /var/lib/apt/lists/*

RUN wget -nc https://github.com/selfhst/icons/archive/refs/heads/main.zip
RUN unzip -jo main.zip icons-main/svg/* -d public/icons/sh
# Install deps & build frontend
RUN bun install && bun run build

# ----------------------------
# Stage 2: Go builder
# ----------------------------
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache bash ca-certificates git && update-ca-certificates

WORKDIR /app

# Copy Go module files first
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Install templ CLI & generate Go code
RUN go install github.com/a-h/templ/cmd/templ@v0.3.865
RUN templ generate -proxy='http://localhost:5173' --path=./internal/ui/

# Copy built frontend from previous stage
COPY --from=frontend /web/dist ./web/dist

# Build Go static binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /tailnetd ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /healthcheck ./cmd/healthcheck/main.go

# ----------------------------
# Stage 3: Final minimal image
# ----------------------------
FROM scratch
LABEL org.opencontainers.image.title="Tailnet" \
      org.opencontainers.image.description="Server for handling tailnets" \
      org.opencontainers.image.version="0.0.1b1" \
      org.opencontainers.image.licenses="AGPL3" \
      org.opencontainers.image.source="https://github.com/sudosu404/tailnet-libs" \
      org.opencontainers.image.authors="Hector <hector@email.gnx> @sudosu404"
# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binaries
COPY --from=builder /tailnetd /tailnetd
COPY --from=builder /healthcheck /healthcheck

# Create minimal passwd entry for non-root user
COPY --from=builder /etc/passwd /etc/passwd

# Writable directories
VOLUME /data /config /tmp
ENV TMPDIR=/tmp

USER 1000

# Expose ports
EXPOSE 8080
EXPOSE 7331

# Healthcheck
HEALTHCHECK CMD ["/healthcheck"]

# Run the main binary
ENTRYPOINT ["/tailnetd"]
