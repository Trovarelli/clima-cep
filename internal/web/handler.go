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

type WeatherUseCase interface {
	Execute(ctx context.Context, zipcode string) (*entity.TemperatureResponse, error)
}

type Handler struct {
	useCase WeatherUseCase
	logger  *slog.Logger
}

func NewHandler(uc WeatherUseCase, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		useCase: uc,
		logger:  logger,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /healthz", h.handleHealth)

	mux.HandleFunc("GET /clima/temp", h.handleWeather)
	mux.HandleFunc("GET /clima/{cep}", h.handleWeather)
	mux.HandleFunc("GET /clima", h.handleWeather)
	mux.HandleFunc("GET /weather/{cep}", h.handleWeather)
	mux.HandleFunc("GET /weather", h.handleWeather)

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
	if r.URL.Query().Has("cep") || r.URL.Query().Has("zipcode") {
		h.handleWeather(w, r)
		return
	}

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
	if cep := strings.TrimSpace(r.URL.Query().Get("cep")); cep != "" {
		return cep
	}
	if zipcode := strings.TrimSpace(r.URL.Query().Get("zipcode")); zipcode != "" {
		return zipcode
	}

	if cep := strings.TrimSpace(r.PathValue("cep")); cep != "" {
		return cep
	}

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
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte("invalid zipcode"))

	case errors.Is(err, usecase.ErrZipcodeNotFound):
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("can not find zipcode"))

	default:
		h.logger.Error("erro interno ao processar requisicao", "erro", err)
		w.WriteHeader(http.StatusInternalServerError)
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
