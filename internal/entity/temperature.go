package entity

import "math"

// TemperatureResponse representa a resposta com as temperaturas nas três escalas (Celsius, Fahrenheit e Kelvin).
type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// ConvertTemperature converte uma temperatura em Celsius para Fahrenheit e Kelvin.
// Fórmulas:
// - Celsius para Fahrenheit: F = C * 1.8 + 32
// - Celsius para Kelvin: K = C + 273.15 (utiliza a constante oficial que atende ao exemplo: 28.5 °C -> 301.65 K)
func ConvertTemperature(celsius float64) TemperatureResponse {
	fahrenheit := celsius*1.8 + 32
	kelvin := celsius + 273.15

	return TemperatureResponse{
		TempC: roundToTwoDecimals(celsius),
		TempF: roundToTwoDecimals(fahrenheit),
		TempK: roundToTwoDecimals(kelvin),
	}
}

func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}
