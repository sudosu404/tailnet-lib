# ----------------------------
# Stage 1: Frontend build environment
# ----------------------------
FROM oven/bun:1 AS frontend

WORKDIR /src
COPY web ./web

# Install frontend deps, download icons, build
RUN bun install --cwd web \
 && mkdir -p web/public/icons/sh \
 && wget -nc https://github.com/selfhst/icons/archive/refs/heads/main.zip -P web \
 && unzip -jo web/main.zip web/icons-main/svg/* -d web/public/icons/sh \
 && rm -f web/main.zip \
 && bun run build --cwd web

# ----------------------------
# Stage 2: Go builder
# ----------------------------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache bash git ca-certificates unzip wget \
    && update-ca-certificates

WORKDIR /app

# Copy Go module files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Copy prebuilt frontend assets
COPY --from=frontend /src/web/dist ./web/dist
COPY --from=frontend /src/web/node_modules ./web/node_modules
COPY --from=frontend /src/web/public/icons ./web/public/icons

# Run go generate if you need (optional, no Bun commands inside)
RUN go generate ./web || true

# Install templ CLI & generate backend UI templates
RUN go install github.com/a-h/templ/cmd/templ@v0.3.865
RUN templ generate -proxy='http://localhost:5173' --path=./internal/ui/

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

# Copy CA certificates and binaries
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /tailnetd /tailnetd
COPY --from=builder /healthcheck /healthcheck
COPY --from=builder /etc/passwd /etc/passwd

VOLUME /data /config /tmp
ENV TMPDIR=/tmp

USER 1000

EXPOSE 8080 7331
HEALTHCHECK CMD ["/healthcheck"]
ENTRYPOINT ["/tailnetd"]