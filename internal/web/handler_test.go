package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clima-cep/internal/entity"
	"clima-cep/internal/usecase"
)

type mockUseCase struct {
	executeFunc func(ctx context.Context, zipcode string) (*entity.TemperatureResponse, error)
}

func (m *mockUseCase) Execute(ctx context.Context, zipcode string) (*entity.TemperatureResponse, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, zipcode)
	}
	return nil, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandler_Success_VariousRoutes(t *testing.T) {
	expectedResponse := &entity.TemperatureResponse{
		TempC: 28.5,
		TempF: 83.3,
		TempK: 301.65,
	}

	mockUC := &mockUseCase{
		executeFunc: func(_ context.Context, zipcode string) (*entity.TemperatureResponse, error) {
			if zipcode != "01001000" {
				t.Errorf("CEP recebido inesperado: %s", zipcode)
			}
			return expectedResponse, nil
		},
	}

	handler := NewHandler(mockUC, testLogger()).Routes()

	routes := []string{
		"/clima/01001000",
		"/01001000",
		"/clima?cep=01001000",
		"/?cep=01001000",
		"/clima/temp?cep=01001000",
		"/weather/01001000",
		"/weather?cep=01001000",
		"/clima?zipcode=01001000",
	}

	for _, route := range routes {
		t.Run("Route "+route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("Esperava status 200, obteve %d", rec.Code)
			}

			if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
				t.Errorf("Esperava Content-Type application/json, obteve %s", rec.Header().Get("Content-Type"))
			}

			var body entity.TemperatureResponse
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("Falha ao decodificar JSON: %v", err)
			}

			if body.TempC != 28.5 || body.TempF != 83.3 || body.TempK != 301.65 {
				t.Errorf("Valores inesperados no corpo: %+v", body)
			}
		})
	}
}

func TestHandler_InvalidZipcode_422(t *testing.T) {
	mockUC := &mockUseCase{
		executeFunc: func(_ context.Context, _ string) (*entity.TemperatureResponse, error) {
			return nil, usecase.ErrInvalidZipcode
		},
	}

	handler := NewHandler(mockUC, testLogger()).Routes()

	routes := []string{
		"/clima/123",
		"/1234567",
		"/clima?cep=inválido",
		"/?cep=",
		"/clima/temp?cep=01001-000",
	}

	for _, route := range routes {
		t.Run("Route "+route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("Esperava status 422, obteve %d", rec.Code)
			}

			if rec.Body.String() != "invalid zipcode" {
				t.Errorf("Esperava corpo 'invalid zipcode', obteve %q", rec.Body.String())
			}
		})
	}
}

func TestHandler_ZipcodeNotFound_404(t *testing.T) {
	mockUC := &mockUseCase{
		executeFunc: func(_ context.Context, _ string) (*entity.TemperatureResponse, error) {
			return nil, usecase.ErrZipcodeNotFound
		},
	}

	handler := NewHandler(mockUC, testLogger()).Routes()

	req := httptest.NewRequest(http.MethodGet, "/clima/99999999", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("Esperava status 404, obteve %d", rec.Code)
	}

	if rec.Body.String() != "can not find zipcode" {
		t.Errorf("Esperava corpo 'can not find zipcode', obteve %q", rec.Body.String())
	}
}

func TestHandler_InternalServerError_500(t *testing.T) {
	mockUC := &mockUseCase{
		executeFunc: func(_ context.Context, _ string) (*entity.TemperatureResponse, error) {
			return nil, errors.New("upstream service failed")
		},
	}

	handler := NewHandler(mockUC, testLogger()).Routes()

	req := httptest.NewRequest(http.MethodGet, "/clima/01001000", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Esperava status 500, obteve %d", rec.Code)
	}

	if rec.Body.String() != "internal server error" {
		t.Errorf("Esperava corpo 'internal server error', obteve %q", rec.Body.String())
	}
}

func TestHandler_HealthEndpoints(t *testing.T) {
	handler := NewHandler(&mockUseCase{}, testLogger()).Routes()

	for _, endpoint := range []string{"/health", "/healthz"} {
		t.Run("Health "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("Esperava status 200, obteve %d", rec.Code)
			}

			if rec.Body.String() != `{"status":"ok"}` {
				t.Errorf("Esperava status ok, obteve %s", rec.Body.String())
			}
		})
	}
}

func TestHandler_Root_WithoutParams_ReturnsUsage(t *testing.T) {
	handler := NewHandler(&mockUseCase{}, testLogger()).Routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Esperava status 200, obteve %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "Sistema de Clima por CEP") {
		t.Errorf("Esperava mensagem de instrução, obteve: %s", rec.Body.String())
	}
}
