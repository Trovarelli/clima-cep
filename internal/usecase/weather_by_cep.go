package usecase

import (
	"context"
	"errors"
	"fmt"

	"clima-cep/internal/entity"
	"clima-cep/internal/infra/viacep"
	"clima-cep/internal/infra/weatherapi"
)

var (
	// ErrInvalidZipcode é retornado quando o formato do CEP é inválido (não possui 8 dígitos numéricos).
	ErrInvalidZipcode = errors.New("invalid zipcode")
	// ErrZipcodeNotFound é retornado quando o CEP ou a localidade não são encontrados.
	ErrZipcodeNotFound = errors.New("can not find zipcode")
)

// ViaCEPClient define a interface necessária para buscar a localização por CEP.
type ViaCEPClient interface {
	GetLocation(ctx context.Context, zipcode string) (*viacep.Location, error)
}

// WeatherClient define a interface necessária para buscar a temperatura por cidade.
type WeatherClient interface {
	GetCurrentTemperatureCelsius(ctx context.Context, city string) (float64, error)
}

// WeatherByCEPUseCase orquestra a validação do CEP, busca da localização e obtenção das temperaturas.
type WeatherByCEPUseCase struct {
	viaCEPClient  ViaCEPClient
	weatherClient WeatherClient
}

// NewWeatherByCEPUseCase cria uma nova instância do caso de uso.
func NewWeatherByCEPUseCase(viaCEP ViaCEPClient, weather WeatherClient) *WeatherByCEPUseCase {
	return &WeatherByCEPUseCase{
		viaCEPClient:  viaCEP,
		weatherClient: weather,
	}
}

// Execute executa o fluxo completo:
// 1. Validação do formato do CEP (8 dígitos numéricos)
// 2. Consulta de endereço no ViaCEP
// 3. Consulta de temperatura na WeatherAPI
// 4. Conversão para Celsius, Fahrenheit e Kelvin
func (uc *WeatherByCEPUseCase) Execute(ctx context.Context, zipcode string) (*entity.TemperatureResponse, error) {
	if !entity.IsValidZipcode(zipcode) {
		return nil, ErrInvalidZipcode
	}

	location, err := uc.viaCEPClient.GetLocation(ctx, zipcode)
	if err != nil {
		if errors.Is(err, viacep.ErrZipcodeNotFound) {
			return nil, ErrZipcodeNotFound
		}
		return nil, fmt.Errorf("falha ao consultar localizacao: %w", err)
	}

	tempC, err := uc.weatherClient.GetCurrentTemperatureCelsius(ctx, location.City)
	if err != nil {
		if errors.Is(err, weatherapi.ErrLocationNotFound) {
			return nil, ErrZipcodeNotFound
		}
		return nil, fmt.Errorf("falha ao consultar clima: %w", err)
	}

	resp := entity.ConvertTemperature(tempC)
	return &resp, nil
}
