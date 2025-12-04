package usecase

import (
	"context"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/repository"
)

type WeatherUseCase interface {
	GetWeatherByZipcode(ctx context.Context, zipcode string) (*domain.Weather, error)
}

type weatherUseCase struct {
	viacepClient  repository.ViaCEPClient
	weatherClient repository.WeatherClient
}

func NewWeatherUseCase(viacepClient repository.ViaCEPClient, weatherClient repository.WeatherClient) WeatherUseCase {
	return &weatherUseCase{
		viacepClient:  viacepClient,
		weatherClient: weatherClient,
	}
}

func (u *weatherUseCase) GetWeatherByZipcode(ctx context.Context, zipcode string) (*domain.Weather, error) {
	// 1. Buscar localização pelo CEP
	location, err := u.viacepClient.GetLocationByZipcode(ctx, zipcode)
	if err != nil {
		return nil, err
	}

	// 2. Buscar temperatura pela localização
	tempCelsius, err := u.weatherClient.GetTemperatureByLocation(ctx, location)
	if err != nil {
		return nil, err
	}

	// 3. Criar objeto Weather com conversões e cidade
	weather := domain.NewWeather(location.City, tempCelsius)

	return &weather, nil
}
