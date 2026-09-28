package weatherapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetCurrentTemperatureCelsius_MissingAPIKey(t *testing.T) {
	client := NewClient(nil, "", "")
	_, err := client.GetCurrentTemperatureCelsius(context.Background(), "São Paulo")
	if !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("Esperava ErrMissingAPIKey, obteve: %v", err)
	}
}

func TestClient_GetCurrentTemperatureCelsius_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/current.json" {
			t.Errorf("Path inesperado: %s", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "dummy-key" {
			t.Errorf("Chave de API inesperada: %s", r.URL.Query().Get("key"))
		}
		if r.URL.Query().Get("q") != "São Paulo" {
			t.Errorf("Query q inesperada: %s", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"location": {
				"name": "São Paulo",
				"country": "Brazil"
			},
			"current": {
				"temp_c": 28.5
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "dummy-key")
	temp, err := client.GetCurrentTemperatureCelsius(context.Background(), "São Paulo")
	if err != nil {
		t.Fatalf("Esperava sucesso, obteve: %v", err)
	}

	if temp != 28.5 {
		t.Errorf("Temperatura esperada 28.5, obteve: %v", temp)
	}
}

func TestClient_GetCurrentTemperatureCelsius_LocationNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"error": {
				"code": 1006,
				"message": "No matching location found."
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "dummy-key")
	_, err := client.GetCurrentTemperatureCelsius(context.Background(), "CidadeInexistente")
	if !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("Esperava ErrLocationNotFound, obteve: %v", err)
	}
}

func TestClient_GetCurrentTemperatureCelsius_InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "dummy-key")
	_, err := client.GetCurrentTemperatureCelsius(context.Background(), "São Paulo")
	if !errors.Is(err, ErrUpstreamError) {
		t.Fatalf("Esperava ErrUpstreamError, obteve: %v", err)
	}
}

func TestClient_GetCurrentTemperatureCelsius_MissingTempC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"current": {}}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "dummy-key")
	_, err := client.GetCurrentTemperatureCelsius(context.Background(), "São Paulo")
	if !errors.Is(err, ErrUpstreamError) {
		t.Fatalf("Esperava ErrUpstreamError para temp_c ausente, obteve: %v", err)
	}
}

func TestClient_GetCurrentTemperatureCelsius_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL, "dummy-key")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetCurrentTemperatureCelsius(ctx, "São Paulo")
	if err == nil {
		t.Fatal("Esperava erro de timeout/cancelamento, obteve nil")
	}
}
