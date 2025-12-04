package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/pkg/otel"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/handler"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/repository"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-b/internal/usecase"
	"github.com/spf13/viper"
)

func main() {
	config := setupConfig()

	// Inicializa o OpenTelemetry
	shutdown, err := otel.InitTracer("service-b", config.GetString("zipkin_url"))
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer shutdown()

	// Configura os clientes
	viacepClient := repository.NewViaCEPClient(config.GetString("viacep_base_url"))
	weatherClient := repository.NewWeatherClient(config.GetString("weather_api_base_url"), config.GetString("weather_api_key"))
	weatherUseCase := usecase.NewWeatherUseCase(viacepClient, weatherClient)
	weatherHandler := handler.NewWeatherHandler(weatherUseCase)

	// Configura o servidor
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.GetInt("port")),
		Handler:      weatherHandler.SetupRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Inicia o servidor
	go func() {
		log.Printf("Service B running on port %d", config.GetInt("port"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Aguarda o sinal de parada
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Stopping Service B...")

	// Fecha o servidor
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Encerra o servidor
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error stopping server: %v", err)
	}

	log.Println("Service B stopped successfully")
}

func setupConfig() *viper.Viper {
	v := viper.New()

	v.SetDefault("port", 8081)
	v.SetDefault("weather_api_key", "")
	v.SetDefault("weather_api_base_url", "https://api.weatherapi.com/v1")
	v.SetDefault("viacep_base_url", "https://viacep.com.br/ws")
	v.SetDefault("zipkin_url", "http://localhost:9411/api/v2/spans")

	v.AutomaticEnv()

	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
		log.Println("Using default values and environment variables")
	} else {
		log.Println(".env file loaded")
	}

	if v.GetString("weather_api_key") == "" {
		log.Fatal("WEATHER_API_KEY is required. Configure in .env or as environment variable")
	}

	return v
}
