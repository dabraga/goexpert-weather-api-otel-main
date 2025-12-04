package usecase

import (
	"context"
	"testing"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockViaCEPClient struct {
	mock.Mock
}

func (m *MockViaCEPClient) GetLocationByZipcode(ctx context.Context, zipcode string) (*domain.Location, error) {
	args := m.Called(ctx, zipcode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Location), args.Error(1)
}

type MockWeatherClient struct {
	mock.Mock
}

func (m *MockWeatherClient) GetTemperatureByLocation(ctx context.Context, location *domain.Location) (float64, error) {
	args := m.Called(ctx, location)
	return args.Get(0).(float64), args.Error(1)
}

func TestWeatherUseCaseGetWeatherByZipcode(t *testing.T) {
	tests := []struct {
		name               string
		zipcode            string
		mockLocation       *domain.Location
		mockLocationErr    error
		mockTemperature    float64
		mockTemperatureErr error
		expectedWeather    *domain.Weather
		expectedErr        error
	}{
		{
			name:    "success - valid zipcode",
			zipcode: "26140040",
			mockLocation: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockLocationErr:    nil,
			mockTemperature:    25.5,
			mockTemperatureErr: nil,
			expectedWeather: &domain.Weather{
				City:  "Belford Roxo",
				TempC: 25.5,
				TempF: 77.9,
				TempK: 298.5,
			},
			expectedErr: nil,
		},
		{
			name:               "error - invalid zipcode",
			zipcode:            "123",
			mockLocation:       nil,
			mockLocationErr:    domain.ErrInvalidZipcode,
			mockTemperature:    0,
			mockTemperatureErr: nil,
			expectedWeather:    nil,
			expectedErr:        domain.ErrInvalidZipcode,
		},
		{
			name:               "error - zipcode not found",
			zipcode:            "99999999",
			mockLocation:       nil,
			mockLocationErr:    domain.ErrZipcodeNotFound,
			mockTemperature:    0,
			mockTemperatureErr: nil,
			expectedWeather:    nil,
			expectedErr:        domain.ErrZipcodeNotFound,
		},
		{
			name:    "error - invalid location",
			zipcode: "26140040",
			mockLocation: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockLocationErr:    nil,
			mockTemperature:    0,
			mockTemperatureErr: domain.ErrInvalidLocation,
			expectedWeather:    nil,
			expectedErr:        domain.ErrInvalidLocation,
		},
		{
			name:    "erro - clima não encontrado",
			zipcode: "26140040",
			mockLocation: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockLocationErr:    nil,
			mockTemperature:    0,
			mockTemperatureErr: domain.ErrWeatherNotFound,
			expectedWeather:    nil,
			expectedErr:        domain.ErrWeatherNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar mocks
			mockViaCEP := new(MockViaCEPClient)
			mockWeather := new(MockWeatherClient)

			// Configurar expectativas dos mocks
			mockViaCEP.On("GetLocationByZipcode", mock.Anything, tt.zipcode).Return(tt.mockLocation, tt.mockLocationErr)
			if tt.mockLocation != nil {
				mockWeather.On("GetTemperatureByLocation", mock.Anything, tt.mockLocation).Return(tt.mockTemperature, tt.mockTemperatureErr)
			}

			// Criar usecase com mocks
			usecase := NewWeatherUseCase(mockViaCEP, mockWeather)

			// Executar teste
			result, err := usecase.GetWeatherByZipcode(context.Background(), tt.zipcode)

			// Verificar resultado
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWeather, result)
			}

			// Verificar se todos os mocks foram chamados
			mockViaCEP.AssertExpectations(t)
			if tt.mockLocation != nil {
				mockWeather.AssertExpectations(t)
			}
		})
	}
}
