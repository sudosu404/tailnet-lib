
# Usa uma imagem oficial do Go como base para a compilação
FROM golang:1.24 AS builder
RUN apk add --no-cache ca-certificates && update-ca-certificates 2>/dev/null || true

# Define o diretório de trabalho
WORKDIR /app

# Copia o código fonte para o container
COPY . .

# Compila a aplicação Go
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /tailnetd ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /healthcheck ./cmd/healthcheck/main.go


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /tailnetd /tailnetd
COPY --from=builder /healthcheck /healthcheck

ENTRYPOINT ["/tailnetd"]

EXPOSE 8080
HEALTHCHECK CMD [ "/healthcheck" ]
