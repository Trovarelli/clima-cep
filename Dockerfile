# Estágio de compilação (Builder)
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instala certificados CA necessários para chamadas HTTPS externas (ViaCEP e WeatherAPI)
RUN apk --no-cache add ca-certificates tzdata

# Gerenciamento de dependências
COPY go.mod ./
RUN go mod download

# Copia todo o código fonte
COPY . .

# Compila o binário estático e otimizado para Linux AMD64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/server

# Estágio final (imagem leve e segura)
FROM alpine:3.21

# Adiciona certificados CA e cria usuário não-privilegiado
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 10001 appgroup && \
    adduser -D -u 10001 -G appgroup -s /sbin/nologin appuser

WORKDIR /app

# Copia o binário compilado
COPY --from=builder /app/server /app/server

# Executa com usuário não-root por segurança
USER appuser:appgroup

# Porta padrão utilizada pelo Google Cloud Run
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
