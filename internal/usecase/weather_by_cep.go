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
	ErrInvalidZipcode  = errors.New("invalid zipcode")
	ErrZipcodeNotFound = errors.New("can not find zipcode")
)

type ViaCEPClient interface {
	GetLocation(ctx context.Context, zipcode string) (*viacep.Location, error)
}

type WeatherClient interface {
	GetCurrentTemperatureCelsius(ctx context.Context, city string) (float64, error)
}

type WeatherByCEPUseCase struct {
	viaCEPClient  ViaCEPClient
	weatherClient WeatherClient
}

func NewWeatherByCEPUseCase(viaCEP ViaCEPClient, weather WeatherClient) *WeatherByCEPUseCase {
	return &WeatherByCEPUseCase{
		viaCEPClient:  viaCEP,
		weatherClient: weather,
	}
}

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
