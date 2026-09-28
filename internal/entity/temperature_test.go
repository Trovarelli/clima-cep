package entity

import (
	"math"
	"testing"
)

func TestConvertTemperature_OfficialExample(t *testing.T) {
	// Cenário 1 oficial: 28.5 °C -> 83.3 °F e 301.65 K
	got := ConvertTemperature(28.5)

	if got.TempC != 28.5 {
		t.Errorf("TempC = %v, esperado 28.5", got.TempC)
	}
	if got.TempF != 83.3 {
		t.Errorf("TempF = %v, esperado 83.3", got.TempF)
	}
	if got.TempK != 301.65 {
		t.Errorf("TempK = %v, esperado 301.65", got.TempK)
	}
}

func TestConvertTemperature_VariousValues(t *testing.T) {
	tests := []struct {
		name      string
		celsius   float64
		wantF     float64
		wantK     float64
	}{
		{
			name:    "Zero absoluto em Celsius (0 °C)",
			celsius: 0.0,
			wantF:   32.0,
			wantK:   273.15,
		},
		{
			name:    "Ponto de ebulição da água (100 °C)",
			celsius: 100.0,
			wantF:   212.0,
			wantK:   373.15,
		},
		{
			name:    "Temperatura negativa (-10 °C)",
			celsius: -10.0,
			wantF:   14.0,
			wantK:   263.15,
		},
		{
			name:    "Valor fracionário (21.2 °C)",
			celsius: 21.2,
			wantF:   70.16,
			wantK:   294.35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertTemperature(tt.celsius)

			if math.Abs(got.TempC-tt.celsius) > 1e-6 {
				t.Errorf("TempC = %v, esperado %v", got.TempC, tt.celsius)
			}
			if math.Abs(got.TempF-tt.wantF) > 1e-6 {
				t.Errorf("TempF = %v, esperado %v", got.TempF, tt.wantF)
			}
			if math.Abs(got.TempK-tt.wantK) > 1e-6 {
				t.Errorf("TempK = %v, esperado %v", got.TempK, tt.wantK)
			}
		})
	}
}
