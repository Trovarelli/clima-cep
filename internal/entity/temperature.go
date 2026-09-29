package entity

import "math"

type TemperatureResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

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
