# ==============================================================================
# Variáveis
# ==============================================================================
SERVICE_A_PORT?=8080
SERVICE_B_PORT?=8081

# Cores
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m

# ==============================================================================
# Comandos Principais
# ==============================================================================
.PHONY: setup run test docker help

setup: ## Configura o ambiente
	@echo "$(BLUE)🔧 Configurando ambiente...$(NC)"
	@cp service-a/.env.example service-a/.env
	@cp service-b/.env.example service-b/.env
	@cd service-a && go mod download && go mod tidy
	@cd service-b && go mod download && go mod tidy
	@echo "$(GREEN)✅ Ambiente configurado!$(NC)"
	@echo "$(YELLOW)📝 Configure WEATHER_API_KEY em service-b/.env$(NC)"

run: ## Roda ambos os serviços
	@echo "$(BLUE)🚀 Iniciando serviços...$(NC)"
	@echo "$(YELLOW)Service A: http://localhost:8080$(NC)"
	@echo "$(YELLOW)Service B: http://localhost:8081$(NC)"
	@echo "$(YELLOW)Zipkin: http://localhost:9411$(NC)"
	@docker-compose up --build

test: ## Roda todos os testes
	@echo "$(BLUE)🧪 Executando testes...$(NC)"
	@cd service-a && go test -v ./...
	@cd service-b && go test -v ./...

docker: ## Comandos Docker
	@echo "$(BLUE)🐳 Comandos Docker:$(NC)"
	@echo "  make docker-up    - Sobe a stack"
	@echo "  make docker-down  - Para a stack"
	@echo "  make docker-logs  - Mostra logs"

docker-up: ## Sobe a stack
	@docker-compose up --build

docker-down: ## Para a stack
	@docker-compose down

docker-logs: ## Mostra logs
	@docker-compose logs -f

help: ## Mostra ajuda
	@echo "$(BLUE)Comandos disponíveis:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

.DEFAULT_GOAL := help