# ADR-0001: Arquitetura de Microserviços com OpenTelemetry

## Status

Aceito

## Contexto

O projeto original era um sistema monolítico que recebia um CEP, buscava a localização no ViaCEP e a temperatura na WeatherAPI. Para demonstrar e estudar observabilidade em microserviços, precisávamos:

1. **Observabilidade**: Implementar tracing distribuído para monitorar latência e debug
2. **Simulação de microserviços**: Criar um ecossistema distribuído para demonstrar conceitos de observabilidade
3. **Aprendizado prático**: Vivenciar os desafios e benefícios de sistemas distribuídos
4. **Demonstração técnica**: Mostrar como OpenTelemetry funciona em cenários reais

## Decisão

Implementar uma arquitetura de microserviços com dois serviços:

### Serviço A (Gateway)

- **Responsabilidade**: Validação de entrada e roteamento (simulação de gateway)
- **Endpoint**: `POST /weather` recebendo `{"cep": "12345678"}`
- **Validações**: CEP com 8 dígitos, formato string
- **Ação**: Encaminhar para Serviço B e retornar resposta

### Serviço B (Processador)

- **Responsabilidade**: Toda a lógica de negócio (mantém funcionalidade original)
- **Endpoint**: `POST /weather` recebendo `{"cep": "12345678"}`
- **Integrações**: ViaCEP (localização) + WeatherAPI (temperatura)
- **Ação**: Processar CEP e retornar dados completos

### Observabilidade

- **OpenTelemetry**: Para instrumentação e coleta de traces
- **Zipkin**: Para visualização e análise de traces
- **OTEL Collector**: Para agregação e exportação de dados

## Implementação

### Estrutura de Projeto

```bash
/
├── pkg/otel/              # Código compartilhado
├── service-a/             # Gateway
├── service-b/             # Processador
├── docker-compose.yml     # Stack completa
└── Makefile               # Comandos de desenvolvimento
```

### Tecnologias

- **Go**: Linguagem principal
- **OpenTelemetry**: Instrumentação
- **Zipkin**: Backend de tracing
- **Docker**: Containerização

### Monitoramento

- Traces distribuídos para debug e análise de latência
- Alertas para falhas de comunicação e APIs externas
- Métricas de performance por serviço

## Conclusão

A decisão de implementar microserviços com OpenTelemetry atende ao objetivo principal de demonstrar observabilidade em sistemas distribuídos. A arquitetura escolhida simula um ecossistema real de microserviços, permitindo vivenciar os desafios e benefícios da observabilidade distribuída em um contexto prático e educativo.
