package weatherapi

import (
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

const defaultBaseURL = "https://api.weatherapi.com/v1"

var (
	ErrLocationNotFound = errors.New("can not find zipcode")
	ErrMissingAPIKey    = errors.New("chave da WeatherAPI nao configurada")
	ErrUpstreamError    = errors.New("erro de comunicacao com a weatherapi")
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientInterface interface {
	GetCurrentTemperatureCelsius(ctx context.Context, city string) (float64, error)
}

type Client struct {
	httpClient HTTPClient
	baseURL    string
	apiKey     string
}

func NewClient(httpClient HTTPClient, baseURL, apiKey string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
	}
}

type weatherAPIResponse struct {
	Current struct {
		TempC *float64 `json:"temp_c"`
	} `json:"current"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) GetCurrentTemperatureCelsius(ctx context.Context, city string) (float64, error) {
	if c.apiKey == "" {
		return 0, ErrMissingAPIKey
	}

	endpoint, err := url.JoinPath(c.baseURL, "current.json")
	if err != nil {
		return 0, fmt.Errorf("construcao de url da weatherapi: %w", err)
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return 0, fmt.Errorf("parse da url da weatherapi: %w", err)
	}

	query := u.Query()
	query.Set("key", c.apiKey)
	query.Set("q", city)
	query.Set("aqi", "no")
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, fmt.Errorf("criacao da requisicao weatherapi: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUpstreamError, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("%w: falha ao ler corpo: %v", ErrUpstreamError, err)
	}

	var res weatherAPIResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return 0, fmt.Errorf("%w: falha ao decodificar json: %v", ErrUpstreamError, err)
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusNotFound {
		if res.Error != nil && res.Error.Code == 1006 {
			return 0, ErrLocationNotFound
		}
		return 0, fmt.Errorf("%w: erro %d: %s", ErrUpstreamError, resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w: status %d", ErrUpstreamError, resp.StatusCode)
	}

	if res.Current.TempC == nil {
		return 0, fmt.Errorf("%w: campo temp_c nao presente na resposta", ErrUpstreamError)
	}

	return *res.Current.TempC, nil
}
