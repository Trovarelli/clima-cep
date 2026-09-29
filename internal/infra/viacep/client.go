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
	ErrZipcodeNotFound = errors.New("can not find zipcode")
	ErrUpstreamError   = errors.New("erro de comunicacao com o viacep")
)

type Location struct {
	City  string `json:"city"`
	State string `json:"state"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientInterface interface {
	GetLocation(ctx context.Context, zipcode string) (*Location, error)
}

type Client struct {
	httpClient HTTPClient
	baseURL    string
}

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
