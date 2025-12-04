package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/dto"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/handler"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/repository"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/usecase"

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

func TestWeatherAPIIntegration(t *testing.T) {
	tests := []struct {
		name               string
		zipcode            string
		mockLocation       *domain.Location
		mockLocationErr    error
		mockTemperature    float64
		mockTemperatureErr error
		expectedStatus     int
		expectedResponse   interface{}
	}{
		{
			name:    "success - valid zipcode with weather",
			zipcode: "26140040",
			mockLocation: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockLocationErr:    nil,
			mockTemperature:    25.5,
			mockTemperatureErr: nil,
			expectedStatus:     http.StatusOK,
			expectedResponse: domain.Weather{
				City:  "Belford Roxo",
				TempC: 25.5,
				TempF: 77.9,
				TempK: 298.5,
			},
		},
		{
			name:               "error - invalid zipcode",
			zipcode:            "123",
			mockLocation:       nil,
			mockLocationErr:    domain.ErrInvalidZipcode,
			mockTemperature:    0,
			mockTemperatureErr: nil,
			expectedStatus:     http.StatusUnprocessableEntity,
			expectedResponse: dto.ErrorResponse{
				Message: "invalid zipcode",
			},
		},
		{
			name:               "error - zipcode not found",
			zipcode:            "99999999",
			mockLocation:       nil,
			mockLocationErr:    domain.ErrZipcodeNotFound,
			mockTemperature:    0,
			mockTemperatureErr: nil,
			expectedStatus:     http.StatusNotFound,
			expectedResponse: dto.ErrorResponse{
				Message: "can not find zipcode",
			},
		},
		{
			name:    "error - weather not found",
			zipcode: "26140040",
			mockLocation: &domain.Location{
				City:  "Belford Roxo",
				State: "RJ",
			},
			mockLocationErr:    nil,
			mockTemperature:    0,
			mockTemperatureErr: domain.ErrWeatherNotFound,
			expectedStatus:     http.StatusNotFound,
			expectedResponse: dto.ErrorResponse{
				Message: "weather not found",
			},
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

			// Criar dependências com mocks
			weatherUseCase := usecase.NewWeatherUseCase(mockViaCEP, mockWeather)
			weatherHandler := handler.NewWeatherHandler(weatherUseCase)
			router := weatherHandler.SetupRoutes()

			// Criar JSON body
			requestBody := dto.WeatherRequest{CEP: tt.zipcode}
			jsonBody, _ := json.Marshal(requestBody)

			// Criar requisição
			req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// Executar requisição
			router.ServeHTTP(recorder, req)

			// Verificar status code
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Verificar resposta
			if tt.expectedStatus == http.StatusOK {
				var weather domain.Weather
				err := json.Unmarshal(recorder.Body.Bytes(), &weather)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, weather)
			} else {
				var errorResp dto.ErrorResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResponse, errorResp)
			}

			// Verificar headers
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			// Verificar se todos os mocks foram chamados
			mockViaCEP.AssertExpectations(t)
			if tt.mockLocation != nil {
				mockWeather.AssertExpectations(t)
			}
		})
	}
}

func TestWeatherAPIRealIntegration(t *testing.T) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		t.Skip("WEATHER_API_KEY not configured, skipping real integration test")
	}

	// Criar clientes reais
	viacepClient := repository.NewViaCEPClient("https://viacep.com.br/ws")
	weatherClient := repository.NewWeatherClient("https://api.weatherapi.com/v1", apiKey)
	weatherUseCase := usecase.NewWeatherUseCase(viacepClient, weatherClient)
	weatherHandler := handler.NewWeatherHandler(weatherUseCase)
	router := weatherHandler.SetupRoutes()

	// Testar com CEP real conhecido
	zipcode := "26140040" // Av. General Jose Muller, Belford Roxo, RJ
	requestBody := dto.WeatherRequest{CEP: zipcode}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	// Executar requisição
	router.ServeHTTP(recorder, req)

	// Log da resposta para debug
	t.Logf("Status: %d, Body: %s", recorder.Code, recorder.Body.String())

	// Verificar se retornou sucesso ou erro esperado (pode falhar por problemas de rede/API)
	if recorder.Code == http.StatusOK {
		// Verificar se a resposta é um JSON válido de Weather
		var weather domain.Weather
		err := json.Unmarshal(recorder.Body.Bytes(), &weather)
		assert.NoError(t, err)
		assert.Greater(t, weather.TempC, -50.0)         // Temperatura razoável
		assert.Less(t, weather.TempC, 60.0)             // Temperatura razoável
		assert.Greater(t, weather.TempF, weather.TempC) // F sempre maior que C
		assert.Greater(t, weather.TempK, weather.TempC) // K sempre maior que C
	} else {
		// Se falhar, pelo menos verificar que retornou um erro estruturado
		var errorResp dto.ErrorResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
		assert.NoError(t, err)
		assert.NotEmpty(t, errorResp.Message)
		t.Logf("Real integration test failed (may be network/API problem): %s", errorResp.Message)
	}
}
