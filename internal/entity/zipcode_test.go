package entity

import "testing"

func TestIsValidZipcode(t *testing.T) {
	tests := []struct {
		name    string
		zipcode string
		want    bool
	}{
		{name: "CEP válido com 8 dígitos", zipcode: "01001000", want: true},
		{name: "CEP válido com zeros à esquerda", zipcode: "00000000", want: true},
		{name: "CEP com hífen (inválido)", zipcode: "01001-000", want: false},
		{name: "CEP com 7 dígitos (curto)", zipcode: "0100100", want: false},
		{name: "CEP com 9 dígitos (longo)", zipcode: "010010000", want: false},
		{name: "CEP com letras", zipcode: "0100100A", want: false},
		{name: "CEP com espaços", zipcode: "01001 00", want: false},
		{name: "CEP vazio", zipcode: "", want: false},
		{name: "CEP com caracteres especiais", zipcode: "01001#00", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidZipcode(tt.zipcode); got != tt.want {
				t.Errorf("IsValidZipcode(%q) = %v, esperado %v", tt.zipcode, got, tt.want)
			}
		})
	}
}
