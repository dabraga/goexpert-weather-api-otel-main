package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/dto"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ServiceBClient interface {
	GetWeather(ctx context.Context, cep string) (*dto.WeatherResponse, error)
}

type serviceBClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewServiceBClient(baseURL string) ServiceBClient {
	return &serviceBClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   30 * time.Second,
		},
	}
}

func (c *serviceBClient) GetWeather(ctx context.Context, cep string) (*dto.WeatherResponse, error) {

	// Prepara a requisição
	reqBody := dto.WeatherRequest{CEP: cep}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Cria a requisição HTTP
	url := fmt.Sprintf("%s/weather", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Faz a requisição
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling service B: %w", err)
	}
	defer resp.Body.Close()

	// Lê a resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Verifica o status code
	if resp.StatusCode != http.StatusOK {
		var errorResp dto.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, domain.NewServiceError(resp.StatusCode, string(body))
		}
		return nil, domain.NewServiceError(resp.StatusCode, errorResp.Message)
	}

	// Parsea a resposta de sucesso
	var weatherResp dto.WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &weatherResp, nil
}
