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
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/handler"
	"github.com/ElizCarvalho/fc-pos-golang-lab-weather-api-com-otel/service-a/internal/repository"
	"github.com/spf13/viper"
)

func main() {
	config := setupConfig()

	// Inicializa o OpenTelemetry
	shutdown, err := otel.InitTracer("service-a", config.GetString("zipkin_url"))
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer shutdown()

	// Configura os clientes
	serviceBClient := repository.NewServiceBClient(config.GetString("service_b_url"))
	weatherHandler := handler.NewWeatherHandler(serviceBClient)

	// Configura o servidor
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.GetInt("port")),
		Handler:      weatherHandler.SetupRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Inicia o servidor
	go func() {
		log.Printf("Service A running on port %d", config.GetInt("port"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Aguarda o sinal de parada
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Stopping Service A...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error stopping server: %v", err)
	}

	log.Println("Service A stopped successfully")
}

func setupConfig() *viper.Viper {
	v := viper.New()

	v.SetDefault("port", 8080)
	v.SetDefault("service_b_url", "http://localhost:8081")
	v.SetDefault("zipkin_url", "http://localhost:9411/api/v2/spans")

	v.AutomaticEnv()

	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
		log.Println("Using default values and environment variables")
	} else {
		log.Println(".env file loaded")
	}

	return v
}
