package entity

// IsValidZipcode valida se o CEP fornecido possui exatamente 8 dígitos numéricos.
// Retorna false se o CEP contiver caracteres não numéricos (como hífens, pontos ou letras)
// ou se o tamanho for diferente de 8.
func IsValidZipcode(zipcode string) bool {
	if len(zipcode) != 8 {
		return false
	}

	for i := 0; i < len(zipcode); i++ {
		if zipcode[i] < '0' || zipcode[i] > '9' {
			return false
		}
	}

	return true
}
