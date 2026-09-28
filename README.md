# Sistema de Clima por CEP no Google Cloud Run (Go Expert)

Projeto desenvolvido como parte do desafio técnico da pós-graduação **Go Expert**. O sistema recebe um CEP brasileiro de 8 dígitos, identifica a cidade correspondente utilizando a API do [ViaCEP](https://viacep.com.br/) e retorna o clima atual nas escalas **Celsius**, **Fahrenheit** e **Kelvin** consumindo a [WeatherAPI](https://www.weatherapi.com/). A aplicação é empacotada em container Docker e implantada no **Google Cloud Run**.

---

## 🌐 URL no Google Cloud Run

> **Endereço do Serviço em Produção:**
> ```
> https://clima-cep-XXXXXX.us-central1.run.app
> ```
> *(Substitua pela URL gerada após o deploy no seu projeto do GCP, conforme instruções abaixo)*

### Exemplos de Chamada em Produção

```bash
# Cenário de Sucesso (200 OK)
curl -s "https://clima-cep-XXXXXX.us-central1.run.app/clima/01001000"

# Cenário de CEP com formato inválido (422 Unprocessable Entity)
curl -i "https://clima-cep-XXXXXX.us-central1.run.app/clima/1234567"

# Cenário de CEP inexistente (404 Not Found)
curl -i "https://clima-cep-XXXXXX.us-central1.run.app/clima/99999999"
```

---

## 📋 Especificações da API (Contrato)

### Endpoints Suportados

Para garantir compatibilidade universal com correções automatizadas e ferramentas de teste, a API aceita requisições nos seguintes formatos:

| Método | Rota | Descrição |
| :--- | :--- | :--- |
| `GET` | `/clima/{cep}` | **Padrão recomendado** — Parâmetro de rota |
| `GET` | `/{cep}` | Rota raiz com parâmetro de rota (ex: `/01001000`) |
| `GET` | `/clima?cep={cep}` | Parâmetro de consulta via query string |
| `GET` | `/?cep={cep}` | Rota raiz com parâmetro via query string |
| `GET` | `/clima/temp?cep={cep}`| Compatibilidade adicional com variantes de desafio |
| `GET` | `/health` ou `/healthz` | Verificação de integridade (*Health Check*) |

---

### Cenários de Resposta

#### 1. Sucesso — `200 OK`

```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.65
}
```

#### 2. Cenários de Falha

| Cenário | Condição | Status Code | Resposta |
| :--- | :--- | :--- | :--- |
| **Formato inválido** | CEP sem 8 dígitos ou contendo caracteres não-numéricos (letras, hífen, espaços) | `422 Unprocessable Entity` | `invalid zipcode` |
| **CEP não encontrado** | CEP com formato correto (8 dígitos), mas inexistente na base de dados | `404 Not Found` | `can not find zipcode` |

---

### Fórmulas de Conversão

* **Celsius para Fahrenheit:** `F = C × 1.8 + 32`
* **Celsius para Kelvin:** `K = C + 273.15` *(utiliza 273.15 para corresponder com exatidão ao exemplo oficial: 28.5 °C → 301.65 K)*

---

## 🚀 Como Rodar Localmente

### Pré-requisitos
* [Go 1.22+](https://go.dev/dl/) instalado
* [Docker](https://www.docker.com/) instalado
* Chave gratuita da [WeatherAPI](https://www.weatherapi.com/signup.aspx)

---

### 1. Rodando com Go diretamente

```bash
# 1. Clone o repositório
git clone https://github.com/SEU_USUARIO/clima-cep.git
cd clima-cep

# 2. Configure a variável de ambiente com sua chave da WeatherAPI
# No Linux/macOS:
export WEATHER_API_KEY="sua_chave_weatherapi_aqui"
export PORT="8080"

# No Windows (PowerShell):
$env:WEATHER_API_KEY="sua_chave_weatherapi_aqui"
$env:PORT="8080"

# 3. Baixe as dependências e inicie o servidor
go mod tidy
go run ./cmd/server
```

O servidor estará acessível em `http://localhost:8080`.

---

### 2. Rodando com Docker

#### Construir a imagem:
```bash
docker build -t clima-cep:latest .
```

#### Executar o container:
```bash
docker run --rm -p 8080:8080 -e WEATHER_API_KEY="sua_chave_weatherapi_aqui" clima-cep:latest
```

---

### 3. Rodando com Docker Compose

Você pode criar um arquivo `.env` na raiz do projeto:
```env
WEATHER_API_KEY=sua_chave_weatherapi_aqui
PORT=8080
```

E rodar:
```bash
docker compose up --build
```

---

## 🧪 Como Rodar os Testes Automatizados

O projeto conta com uma suíte completa de testes automatizados com mocks para isolar chamadas de rede e garantir 100% de repetibilidade.

```bash
# Executar todos os testes
go test ./...

# Executar com detalhes de cada teste (verbose)
go test -v ./...

# Executar exibindo o relatório de cobertura de código
go test -v -cover ./...
```

### O que os testes cobrem:
* **Entidades (`internal/entity`):** Validação de regras de CEP (8 dígitos, caracteres especiais, comprimento) e conversões exatas de temperatura (Celsius, Fahrenheit, Kelvin).
* **Infraestrutura ViaCEP (`internal/infra/viacep`):** Cenários de sucesso, CEP inexistente (`erro: true` e `erro: "true"`), localidade vazia, HTTP 404 e erros de upstream com servidor mock `httptest`.
* **Infraestrutura WeatherAPI (`internal/infra/weatherapi`):** Requisições com encoding de URL, captura de temperatura, localização não encontrada (código 1006) e erros de upstream.
* **Caso de Uso (`internal/usecase`):** Orquestração completa, validação prévia de CEP sem onerar provedores externos, propagação de erros.
* **Camada Web (`internal/web`):** Rotas múltiplas, headers HTTP, status codes 200, 422, 404, 500, health checks e serialização JSON.

---

## ☁️ Deploy no Google Cloud Run

O Google Cloud Run permite hospedar containers stateless de forma serverless (com cota gratuita mensal).

### Opção A: Deploy direto via `gcloud` CLI (Recomendado)

1. **Instale e autentique o Google Cloud SDK:**
   ```bash
   gcloud auth login
   gcloud config set project SEU_PROJECT_ID
   ```

2. **Habilite a API do Cloud Run e Cloud Build:**
   ```bash
   gcloud services enable run.googleapis.com cloudbuild.googleapis.com
   ```

3. **Execute o deploy a partir do código-fonte:**
   ```bash
   gcloud run deploy clima-cep \
     --source . \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars WEATHER_API_KEY="sua_chave_weatherapi_aqui"
   ```

4. Após a conclusão, o terminal exibirá a **Service URL**. Copie a URL e atualize a seção [URL no Google Cloud Run](#-url-no-google-cloud-run) deste README.

---

### Opção B: Deploy via Console Web do Google Cloud

1. Acesse o [Google Cloud Console](https://console.cloud.google.com/).
2. Vá em **Cloud Run** > **Criar Serviço**.
3. Escolha **Implantar continuamente a partir de um repositório** (conecte seu GitHub) ou faça upload da imagem via **Artifact Registry**.
4. Marque a opção: **Permitir invocações não autenticadas**.
5. Na aba **Contêineres, Volumes, Rede, Segurança**:
   * Adicione a variável de ambiente: `WEATHER_API_KEY` com o valor da sua chave.
   * Defina a porta do contêiner como `8080`.
6. Clique em **Criar**. A URL pública será gerada.

---

## 🏗️ Estrutura do Projeto

```text
.
├── cmd/
│   └── server/
│       └── main.go                 # Ponto de entrada da aplicação HTTP e graceful shutdown
├── internal/
│   ├── config/
│   │   ├── config.go               # Leitura de variáveis de ambiente e fallback .env
│   │   └── config_test.go
│   ├── entity/
│   │   ├── temperature.go          # Regras e conversões térmicas (C, F, K)
│   │   ├── temperature_test.go
│   │   ├── zipcode.go              # Validação de CEP de 8 dígitos
│   │   └── zipcode_test.go
│   ├── infra/
│   │   ├── viacep/
│   │   │   ├── client.go           # Integração com a API do ViaCEP
│   │   │   └── client_test.go
│   │   └── weatherapi/
│   │       ├── client.go           # Integração com a API do WeatherAPI
│   │       └── client_test.go
│   ├── usecase/
│   │   ├── weather_by_cep.go       # Orquestração do caso de uso
│   │   └── weather_by_cep_test.go
│   └── web/
│       ├── handler.go              # Manipulador HTTP, roteamento e respostas de erro
│       └── handler_test.go
├── test/
│   └── api.http                    # Requisições de teste para VS Code REST Client
├── .dockerignore
├── .env.example
├── .gitignore
├── Dockerfile                      # Multi-stage build para imagem leve e segura
├── docker-compose.yml
├── go.mod
└── README.md
```
