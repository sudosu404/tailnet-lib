FROM golang:1.25-alpine AS builder
RUN apk add --no-cache bash ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy prebuilt frontend dist
COPY web/dist ./web/dist
# Generate templates and build backend
RUN go install github.com/a-h/templ/cmd/templ@v0.3.865
RUN templ generate --path=./internal/ui/
RUN CGO_ENABLED=0 GOOS=linux go build -o /tailnetd ./cmd/server/main.go

FROM scratch
COPY --from=builder /tailnetd /tailnetd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
USER 1000
EXPOSE 8080 7331
ENTRYPOINT ["/tailnetd"]