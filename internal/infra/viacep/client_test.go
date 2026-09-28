package viacep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetLocation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ws/01001000/json" {
			t.Errorf("Path inesperado: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"cep": "01001-000",
			"logradouro": "Praça da Sé",
			"bairro": "Sé",
			"localidade": "São Paulo",
			"uf": "SP"
		}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	loc, err := client.GetLocation(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("Esperava sucesso, obteve erro: %v", err)
	}

	if loc.City != "São Paulo" {
		t.Errorf("City esperada 'São Paulo', obteve %q", loc.City)
	}
	if loc.State != "SP" {
		t.Errorf("State esperado 'SP', obteve %q", loc.State)
	}
}

func TestClient_GetLocation_NotFound_BoolError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"erro": true}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	loc, err := client.GetLocation(context.Background(), "99999999")
	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava ErrZipcodeNotFound, obteve: %v", err)
	}
	if loc != nil {
		t.Fatalf("Esperava Location nil, obteve: %v", loc)
	}
}

func TestClient_GetLocation_NotFound_StringError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"erro": "true"}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	loc, err := client.GetLocation(context.Background(), "99999999")
	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava ErrZipcodeNotFound, obteve: %v", err)
	}
	if loc != nil {
		t.Fatalf("Esperava Location nil, obteve: %v", loc)
	}
}

func TestClient_GetLocation_NotFound_HTTP404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	_, err := client.GetLocation(context.Background(), "99999999")
	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava ErrZipcodeNotFound, obteve: %v", err)
	}
}

func TestClient_GetLocation_EmptyCity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"localidade": "", "uf": "SP"}`))
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	_, err := client.GetLocation(context.Background(), "12345678")
	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava ErrZipcodeNotFound para localidade vazia, obteve: %v", err)
	}
}

func TestClient_GetLocation_InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	_, err := client.GetLocation(context.Background(), "01001000")
	if !errors.Is(err, ErrUpstreamError) {
		t.Fatalf("Esperava ErrUpstreamError, obteve: %v", err)
	}
}

func TestClient_GetLocation_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetLocation(ctx, "01001000")
	if err == nil {
		t.Fatal("Esperava erro de timeout/cancelamento, obteve nil")
	}
}
