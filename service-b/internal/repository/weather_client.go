package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/dto"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type WeatherClient interface {
	GetTemperatureByLocation(ctx context.Context, location *domain.Location) (float64, error)
}

type weatherClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	tracer     trace.Tracer
}

func NewWeatherClient(baseURL, apiKey string) WeatherClient {
	return &weatherClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		tracer: otel.Tracer("service-b"),
	}
}

// GetTemperatureByLocation busca a temperatura pela localização
func (c *weatherClient) GetTemperatureByLocation(ctx context.Context, location *domain.Location) (float64, error) {
	// Criar span para medir tempo da chamada WeatherAPI
	ctx, span := c.tracer.Start(ctx, "service-b.fetch-weather")
	defer span.End()

	if location == nil || location.City == "" {
		span.RecordError(domain.ErrInvalidLocation)
		return 0, domain.ErrInvalidLocation
	}

	query := fmt.Sprintf("%s, %s, Brazil", location.City, location.State)
	requestURL := fmt.Sprintf("%s/current.json?key=%s&q=%s&aqi=no", c.baseURL, c.apiKey, url.QueryEscape(query))

	// Criar requisição com contexto
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("error querying WeatherAPI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		span.RecordError(fmt.Errorf("invalid API key"))
		return 0, fmt.Errorf("invalid API key")
	}
	if resp.StatusCode == http.StatusBadRequest {
		span.RecordError(domain.ErrWeatherNotFound)
		return 0, domain.ErrWeatherNotFound
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("error in WeatherAPI: status %d", resp.StatusCode)
		span.RecordError(err)
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("error reading WeatherAPI response: %w", err)
	}

	var weatherResp dto.WeatherAPIResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("error parsing WeatherAPI response: %w", err)
	}

	return weatherResp.Current.TempC, nil
}
