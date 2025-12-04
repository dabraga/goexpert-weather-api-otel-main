package domain

import (
	"regexp"
)

func ValidateZipcode(cep string) error {
	if len(cep) != 8 {
		return ErrInvalidZipcode
	}

	matched, err := regexp.MatchString(`^\d{8}$`, cep)
	if err != nil {
		return err
	}

	if !matched {
		return ErrInvalidZipcode
	}

	return nil
}
