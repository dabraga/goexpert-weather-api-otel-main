package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWeather(t *testing.T) {
	tests := []struct {
		name        string
		city        string
		tempCelsius float64
		expected    Weather
	}{
		{
			name:        "positive temperature",
			city:        "São Paulo",
			tempCelsius: 25.0,
			expected: Weather{
				City:  "São Paulo",
				TempC: 25.0,
				TempF: 77.0,
				TempK: 298.0,
			},
		},
		{
			name:        "negative temperature",
			city:        "Belford Roxo",
			tempCelsius: -10.0,
			expected: Weather{
				City:  "Belford Roxo",
				TempC: -10.0,
				TempF: 14.0,
				TempK: 263.0,
			},
		},
		{
			name:        "temperature with decimal",
			city:        "Rio de Janeiro",
			tempCelsius: 28.5,
			expected: Weather{
				City:  "Rio de Janeiro",
				TempC: 28.5,
				TempF: 83.3,
				TempK: 301.5,
			},
		},
		{
			name:        "zero absolute",
			city:        "Test City",
			tempCelsius: -273.0,
			expected: Weather{
				City:  "Test City",
				TempC: -273.0,
				TempF: -459.4,
				TempK: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewWeather(tt.city, tt.tempCelsius)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		name     string
		celsius  float64
		expected float64
	}{
		{"zero celsius", 0, 32},
		{"25 celsius", 25, 77},
		{"100 celsius", 100, 212},
		{"-40 celsius", -40, -40}, // Ponto onde C e F são iguais
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := celsiusToFahrenheit(tt.celsius)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCelsiusToKelvin(t *testing.T) {
	tests := []struct {
		name     string
		celsius  float64
		expected float64
	}{
		{"zero celsius", 0, 273},
		{"25 celsius", 25, 298},
		{"100 celsius", 100, 373},
		{"zero absoluto", -273, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := celsiusToKelvin(tt.celsius)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateZipcode(t *testing.T) {
	tests := []struct {
		name      string
		zipcode   string
		expectErr bool
	}{
		{"valid zipcode", "26140040", false},
		{"zipcode with hyphen", "01310-100", true},
		{"zipcode too short", "0131010", true},
		{"zipcode too long", "261400401", true},
		{"zipcode with letters", "0131010a", true},
		{"zipcode empty", "", true},
		{"zipcode with spaces", "01310 100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateZipcode(tt.zipcode)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidZipcode, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatZipcode(t *testing.T) {
	tests := []struct {
		name     string
		zipcode  string
		expected string
	}{
		{"zipcode without hyphen", "26140040", "26140-040"},
		{"zipcode with hyphen", "26140-040", "26140-040"},
		{"invalid zipcode", "261400401", "261400401"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatZipcode(tt.zipcode)
			assert.Equal(t, tt.expected, result)
		})
	}
}
