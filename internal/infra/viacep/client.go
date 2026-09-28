package viacep

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://viacep.com.br"

var (
	// ErrZipcodeNotFound é retornado quando o CEP não é encontrado na base de dados do ViaCEP.
	ErrZipcodeNotFound = errors.New("can not find zipcode")
	// ErrUpstreamError é retornado quando ocorre um erro de comunicação ou resposta inválida do ViaCEP.
	ErrUpstreamError = errors.New("erro de comunicacao com o viacep")
)

// Location representa a localidade retornada pelo ViaCEP.
type Location struct {
	City  string `json:"city"`
	State string `json:"state"`
}

// HTTPClient define a interface para o cliente HTTP, facilitando testes com mocks.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientInterface define o contrato para busca de localização por CEP.
type ClientInterface interface {
	GetLocation(ctx context.Context, zipcode string) (*Location, error)
}

// Client é o cliente para consulta à API do ViaCEP.
type Client struct {
	httpClient HTTPClient
	baseURL    string
}

// NewClient cria uma nova instância de Client para o ViaCEP.
func NewClient(httpClient HTTPClient, baseURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

type viaCepResponse struct {
	Localidade string       `json:"localidade"`
	UF         string       `json:"uf"`
	Erro       flexibleBool `json:"erro"`
}

type flexibleBool bool

func (b *flexibleBool) UnmarshalJSON(data []byte) error {
	trimmed := string(bytes.Trim(data, ` "`))
	if trimmed == "true" {
		*b = true
		return nil
	}
	if trimmed == "false" || trimmed == "" {
		*b = false
		return nil
	}
	return nil
}

// GetLocation busca a cidade e estado correspondentes a um CEP de 8 dígitos.
func (c *Client) GetLocation(ctx context.Context, zipcode string) (*Location, error) {
	endpoint, err := url.JoinPath(c.baseURL, "ws", zipcode, "json")
	if err != nil {
		return nil, fmt.Errorf("construcao de url invalida: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("criacao da requisicao viacep: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstreamError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrZipcodeNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUpstreamError, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: falha ao ler corpo: %v", ErrUpstreamError, err)
	}

	var res viaCepResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("%w: falha ao decodificar json: %v", ErrUpstreamError, err)
	}

	if res.Erro {
		return nil, ErrZipcodeNotFound
	}

	if strings.TrimSpace(res.Localidade) == "" {
		return nil, ErrZipcodeNotFound
	}

	return &Location{
		City:  strings.TrimSpace(res.Localidade),
		State: strings.TrimSpace(res.UF),
	}, nil
}
