package entity

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
