package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/domain"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/dto"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type WeatherHandler struct {
	serviceBClient repository.ServiceBClient
	tracer         trace.Tracer
}

func NewWeatherHandler(serviceBClient repository.ServiceBClient) *WeatherHandler {
	return &WeatherHandler{
		serviceBClient: serviceBClient,
		tracer:         otel.Tracer("service-a"),
	}
}

func (h *WeatherHandler) SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Instrumenta com OpenTelemetry
	r.Post("/weather", otelhttp.NewHandler(http.HandlerFunc(h.GetWeather), "service-a.handle-request").ServeHTTP)

	return r
}

// GetWeather busca a temperatura pelo CEP
func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Cria o span para validação
	ctx, span := h.tracer.Start(ctx, "service-a.validate-input")
	defer span.End()

	// Parsea o body
	var req dto.WeatherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Valida o CEP
	if err := domain.ValidateZipcode(req.CEP); err != nil {
		span.RecordError(err)
		h.writeErrorResponse(w, http.StatusUnprocessableEntity, "invalid zipcode")
		return
	}

	span.End() // Fecha o span de validação

	// Cria o span para chamada ao Serviço B
	ctx, span = h.tracer.Start(ctx, "service-a.call-service-b")
	defer span.End()

	// Chama o Serviço B
	weather, err := h.serviceBClient.GetWeather(ctx, req.CEP)
	if err != nil {
		span.RecordError(err)
		log.Printf("Error calling service B: %v", err)
		h.handleServiceError(w, err)
		return
	}

	// Retorna sucesso
	h.writeJSONResponse(w, http.StatusOK, weather)
}

func (h *WeatherHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func (h *WeatherHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResp := dto.ErrorResponse{
		Message: message,
	}
	h.writeJSONResponse(w, statusCode, errorResp)
}

// Trata erros do Service B preservando o status code
func (h *WeatherHandler) handleServiceError(w http.ResponseWriter, err error) {
	var serviceErr *domain.ServiceError
	if ok := errors.As(err, &serviceErr); ok {
		h.writeErrorResponse(w, serviceErr.StatusCode, serviceErr.Message)
		return
	}

	h.writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
}
