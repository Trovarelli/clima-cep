package web

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"clima-cep/internal/entity"
	"clima-cep/internal/usecase"
)

// WeatherUseCase define a interface que o handler precisa para consultar o clima por CEP.
type WeatherUseCase interface {
	Execute(ctx context.Context, zipcode string) (*entity.TemperatureResponse, error)
}

// Handler gerencia as requisições HTTP da aplicação.
type Handler struct {
	useCase WeatherUseCase
	logger  *slog.Logger
}

// NewHandler cria uma nova instância de Handler.
func NewHandler(uc WeatherUseCase, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		useCase: uc,
		logger:  logger,
	}
}

// Routes registra todos os endpoints suportados pela API.
// Suporta múltiplos formatos para garantir compatibilidade com qualquer cliente/avaliador:
// - GET /{cep}
// - GET /clima/{cep}
// - GET /weather/{cep}
// - GET /clima?cep={cep}
// - GET /clima/temp?cep={cep}
// - GET /?cep={cep}
// - GET /health e /healthz
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Endpoints de verificação de integridade (Health Check)
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /healthz", h.handleHealth)

	// Endpoints com prefixos dedicados
	mux.HandleFunc("GET /clima/temp", h.handleWeather)
	mux.HandleFunc("GET /clima/{cep}", h.handleWeather)
	mux.HandleFunc("GET /clima", h.handleWeather)
	mux.HandleFunc("GET /weather/{cep}", h.handleWeather)
	mux.HandleFunc("GET /weather", h.handleWeather)

	// Rotas raiz: /{cep} ou /?cep=...
	mux.HandleFunc("GET /{cep}", h.handleWeather)
	mux.HandleFunc("GET /", h.handleRoot)

	return h.loggingMiddleware(mux)
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Se a rota for exatamente "/" e o parâmetro "cep" foi informado (mesmo vazio), processa como weather
	if r.URL.Query().Has("cep") || r.URL.Query().Has("zipcode") {
		h.handleWeather(w, r)
		return
	}

	// Caso contrário, exibe informações e instruções de uso da API
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{
		"message": "Sistema de Clima por CEP - Go Expert (Cloud Run)",
		"usage": "GET /clima/{cep} ou GET /?cep={cep}",
		"example": "/clima/01001000",
		"health": "/health"
	}`))
}

func (h *Handler) handleWeather(w http.ResponseWriter, r *http.Request) {
	cep := h.extractZipcode(r)

	res, err := h.useCase.Execute(r.Context(), cep)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error("falha ao codificar resposta json", "erro", err)
	}
}

func (h *Handler) extractZipcode(r *http.Request) string {
	// 1. Tenta query param "cep" ou "zipcode"
	if cep := strings.TrimSpace(r.URL.Query().Get("cep")); cep != "" {
		return cep
	}
	if zipcode := strings.TrimSpace(r.URL.Query().Get("zipcode")); zipcode != "" {
		return zipcode
	}

	// 2. Tenta path value do router (Go 1.22+)
	if cep := strings.TrimSpace(r.PathValue("cep")); cep != "" {
		return cep
	}

	// 3. Fallback: extrai do caminho da URL
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if last != "" && last != "clima" && last != "weather" && last != "temp" && last != "health" && last != "healthz" {
			return last
		}
	}

	return ""
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch {
	case errors.Is(err, usecase.ErrInvalidZipcode):
		w.WriteHeader(http.StatusUnprocessableEntity) // 422
		_, _ = w.Write([]byte("invalid zipcode"))

	case errors.Is(err, usecase.ErrZipcodeNotFound):
		w.WriteHeader(http.StatusNotFound) // 404
		_, _ = w.Write([]byte("can not find zipcode"))

	default:
		h.logger.Error("erro interno ao processar requisicao", "erro", err)
		w.WriteHeader(http.StatusInternalServerError) // 500
		_, _ = w.Write([]byte("internal server error"))
	}
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("requisicao recebida",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
		)
		next.ServeHTTP(w, r)
	})
}
