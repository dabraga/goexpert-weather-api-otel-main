package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/dto"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ViaCEPClient interface {
	GetLocationByZipcode(ctx context.Context, zipcode string) (*domain.Location, error)
}

type viacepClient struct {
	baseURL    string
	httpClient *http.Client
	tracer     trace.Tracer
}

func NewViaCEPClient(baseURL string) ViaCEPClient {
	return &viacepClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		tracer: otel.Tracer("service-b"),
	}
}

func (c *viacepClient) GetLocationByZipcode(ctx context.Context, zipcode string) (*domain.Location, error) {
	// Criar span para medir tempo da chamada ViaCEP
	ctx, span := c.tracer.Start(ctx, "service-b.fetch-zipcode")
	defer span.End()

	if err := domain.ValidateZipcode(zipcode); err != nil {
		span.RecordError(err)
		return nil, err
	}

	formattedZipcode := domain.FormatZipcode(zipcode)
	url := fmt.Sprintf("%s/%s/json/", c.baseURL, formattedZipcode)

	// Criar requisição com contexto
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error querying ViaCEP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error reading ViaCEP response: %w", err)
	}

	var viacepResp dto.ViaCEPResponse
	if err := json.Unmarshal(body, &viacepResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("error parsing ViaCEP response: %w", err)
	}

	if viacepResp.Erro == "true" || viacepResp.Localidade == "" {
		span.RecordError(domain.ErrZipcodeNotFound)
		return nil, domain.ErrZipcodeNotFound
	}

	return &domain.Location{
		City:  viacepResp.Localidade,
		State: viacepResp.UF,
	}, nil
}
