package domain

import (
	"fmt"
	"math"
)

type Weather struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

type Location struct {
	City  string `json:"city"`
	State string `json:"state"`
}

// NewWeather cria uma nova instância de Weather a partir da temperatura em Celsius e cidade
func NewWeather(city string, tempCelsius float64) Weather {
	return Weather{
		City:  city,
		TempC: roundToOneDecimal(tempCelsius),
		TempF: roundToOneDecimal(celsiusToFahrenheit(tempCelsius)),
		TempK: roundToOneDecimal(celsiusToKelvin(tempCelsius)),
	}
}

// Fórmula: F = C * 1.8 + 32
func celsiusToFahrenheit(celsius float64) float64 {
	return celsius*1.8 + 32
}

// Fórmula: K = C + 273
func celsiusToKelvin(celsius float64) float64 {
	return celsius + 273
}

func roundToOneDecimal(value float64) float64 {
	return math.Round(value*10) / 10
}

func ValidateZipcode(zipcode string) error {
	if len(zipcode) != 8 {
		return ErrInvalidZipcode
	}

	for _, char := range zipcode {
		if char < '0' || char > '9' {
			return ErrInvalidZipcode
		}
	}

	return nil
}

func FormatZipcode(zipcode string) string {
	if len(zipcode) == 8 {
		return fmt.Sprintf("%s-%s", zipcode[:5], zipcode[5:])
	}
	return zipcode
}
