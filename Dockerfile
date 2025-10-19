# ----------------------------
# Stage 1: Frontend build environment
# ----------------------------
FROM oven/bun:1 AS frontend

WORKDIR /src/web
COPY web/package.json web/bun.lock* ./
RUN bun install
COPY web ./

# Copy icons manually
RUN mkdir -p public/icons/sh \
 && wget -nc https://github.com/selfhst/icons/archive/refs/heads/main.zip -P . \
 && unzip -jo main.zip 'icons-main/svg/*' -d public/icons/sh \
 && rm -f main.zip

# Build frontend
RUN bun run build

# ----------------------------
# Stage 2: Go builder
# ----------------------------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache bash git ca-certificates unzip wget \
 && update-ca-certificates

WORKDIR /app

# Copy Go modules first
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Copy frontend build
COPY --from=frontend /src/web/dist ./web/dist
COPY --from=frontend /src/web/public ./web/public
COPY --from=frontend /src/web/node_modules ./web/node_modules

# Optional go generate (no Bun)
RUN go generate ./web || true

# Install templ & generate templates
RUN go install github.com/a-h/templ/cmd/templ@v0.3.865
RUN templ generate -proxy='http://localhost:5173' --path=./internal/ui/

# Build static binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /tailnetd ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /healthcheck ./cmd/healthcheck/main.go

# ----------------------------
# Stage 3: Minimal final image
# ----------------------------
FROM scratch

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