package repository

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWeatherClient struct {
	mock.Mock
}

func (m *MockWeatherClient) GetTemperatureByLocation(ctx context.Context, location *domain.Location) (float64, error) {
	args := m.Called(ctx, location)
	return args.Get(0).(float64), args.Error(1)
}

func TestWeatherClientGetTemperatureByLocation(t *testing.T) {
	tests := []struct {
		name           string
		location       *domain.Location
		mockResponse   *dto.WeatherAPIResponse
		mockStatusCode int
		expected       float64
		expectedErr    error
	}{
		{
			name: "success - valid location",
			location: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockResponse: &dto.WeatherAPIResponse{
				Current: struct {
					TempC float64 `json:"temp_c"`
					TempF float64 `json:"temp_f"`
				}{
					TempC: 25.5,
					TempF: 77.9,
				},
			},
			mockStatusCode: http.StatusOK,
			expected:       25.5,
			expectedErr:    nil,
		},
		{
			name:           "error - null location",
			location:       nil,
			mockResponse:   nil,
			mockStatusCode: http.StatusOK,
			expected:       0,
			expectedErr:    domain.ErrInvalidLocation,
		},
		{
			name: "error - location not found",
			location: &domain.Location{
				City:  "CidadeInexistente",
				State: "XX",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusBadRequest,
			expected:       0,
			expectedErr:    domain.ErrWeatherNotFound,
		},
		{
			name: "error - invalid API key",
			location: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockResponse:   nil,
			mockStatusCode: http.StatusUnauthorized,
			expected:       0,
			expectedErr:    errors.New("invalid API key"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar servidor mock
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Criar cliente com URL do servidor mock
			client := NewWeatherClient(server.URL, "test-key")

			// Executar teste
			result, err := client.GetTemperatureByLocation(context.Background(), tt.location)

			// Verificar resultado
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Equal(t, 0.0, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
