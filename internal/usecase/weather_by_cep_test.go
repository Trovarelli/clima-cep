package usecase

import (
	"context"
	"errors"
	"testing"

	"clima-cep/internal/infra/viacep"
	"clima-cep/internal/infra/weatherapi"
)

type mockViaCEPClient struct {
	location *viacep.Location
	err      error
	called   bool
}

func (m *mockViaCEPClient) GetLocation(_ context.Context, _ string) (*viacep.Location, error) {
	m.called = true
	return m.location, m.err
}

type mockWeatherClient struct {
	tempC  float64
	err    error
	called bool
}

func (m *mockWeatherClient) GetCurrentTemperatureCelsius(_ context.Context, _ string) (float64, error) {
	m.called = true
	return m.tempC, m.err
}

func TestWeatherByCEPUseCase_Success(t *testing.T) {
	viaCEPMock := &mockViaCEPClient{
		location: &viacep.Location{City: "São Paulo", State: "SP"},
	}
	weatherMock := &mockWeatherClient{
		tempC: 28.5,
	}

	uc := NewWeatherByCEPUseCase(viaCEPMock, weatherMock)
	res, err := uc.Execute(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("Esperava sucesso, obteve erro: %v", err)
	}

	if !viaCEPMock.called {
		t.Error("ViaCEP client deveria ter sido chamado")
	}
	if !weatherMock.called {
		t.Error("Weather client deveria ter sido chamado")
	}

	if res.TempC != 28.5 {
		t.Errorf("TempC esperada 28.5, obteve %v", res.TempC)
	}
	if res.TempF != 83.3 {
		t.Errorf("TempF esperada 83.3, obteve %v", res.TempF)
	}
	if res.TempK != 301.65 {
		t.Errorf("TempK esperada 301.65, obteve %v", res.TempK)
	}
}

func TestWeatherByCEPUseCase_InvalidZipcode_DoesNotCallExternalAPIs(t *testing.T) {
	invalidCases := []string{
		"1234567",
		"123456789",
		"01001-000",
		"01001A00",
		"",
	}

	for _, tc := range invalidCases {
		viaCEPMock := &mockViaCEPClient{}
		weatherMock := &mockWeatherClient{}

		uc := NewWeatherByCEPUseCase(viaCEPMock, weatherMock)
		_, err := uc.Execute(context.Background(), tc)

		if !errors.Is(err, ErrInvalidZipcode) {
			t.Errorf("Para CEP %q esperava ErrInvalidZipcode, obteve %v", tc, err)
		}
		if viaCEPMock.called {
			t.Errorf("ViaCEP não deveria ser chamado para CEP inválido %q", tc)
		}
		if weatherMock.called {
			t.Errorf("Weather client não deveria ser chamado para CEP inválido %q", tc)
		}
	}
}

func TestWeatherByCEPUseCase_ViaCEPNotFound(t *testing.T) {
	viaCEPMock := &mockViaCEPClient{
		err: viacep.ErrZipcodeNotFound,
	}
	weatherMock := &mockWeatherClient{}

	uc := NewWeatherByCEPUseCase(viaCEPMock, weatherMock)
	_, err := uc.Execute(context.Background(), "99999999")

	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava ErrZipcodeNotFound, obteve %v", err)
	}
	if weatherMock.called {
		t.Error("Weather client não deveria ser chamado se o CEP não for encontrado")
	}
}

func TestWeatherByCEPUseCase_WeatherAPINotFound(t *testing.T) {
	viaCEPMock := &mockViaCEPClient{
		location: &viacep.Location{City: "Desconhecida", State: "XX"},
	}
	weatherMock := &mockWeatherClient{
		err: weatherapi.ErrLocationNotFound,
	}

	uc := NewWeatherByCEPUseCase(viaCEPMock, weatherMock)
	_, err := uc.Execute(context.Background(), "12345678")

	if !errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava ErrZipcodeNotFound quando a cidade não for encontrada na WeatherAPI, obteve %v", err)
	}
}

func TestWeatherByCEPUseCase_UpstreamError(t *testing.T) {
	viaCEPMock := &mockViaCEPClient{
		err: errors.New("timeout"),
	}
	weatherMock := &mockWeatherClient{}

	uc := NewWeatherByCEPUseCase(viaCEPMock, weatherMock)
	_, err := uc.Execute(context.Background(), "01001000")

	if err == nil || errors.Is(err, ErrInvalidZipcode) || errors.Is(err, ErrZipcodeNotFound) {
		t.Fatalf("Esperava erro de upstream, obteve: %v", err)
	}
}
