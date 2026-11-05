# API de Pagamentos

API em Go para simular fluxo de autorização, compensação, liquidação e reembolso de transações em cartões virtuais.

## Visão geral

Esta aplicação expõe endpoints REST para:

- autorizar compras com cartão virtual;
- cancelar autorizações;
- processar lote de clearing;
- processar lote de settlement;
- reembolsar transações liquidadas;
- consultar cartões e autorizações;
- monitorar métricas de observabilidade com Prometheus.

A aplicação usa o framework Gin e mantém os dados em memória, o que torna o projeto ideal para demonstrações, testes e exercícios de arquitetura de pagamentos.

## Estrutura do projeto

- `main.go` — inicializa o router e registra as rotas;
- `handlers/handlers.go` — camada HTTP com os endpoints da API;
- `services/card_service.go` — regras de negócio de autorização, clearing, settlement e refund;
- `repository/memory.go` — armazenamento em memória dos cartões e transações;
- `models/models.go` — structs e constantes compartilhadas;
- `observability/metrics.go` — métricas Prometheus e middleware de HTTP;
- `healthcheck/healthcheck.go` — endpoint de saúde;
- `docker-compose.yaml` — ambiente com aplicação, Prometheus e Grafana;
- `Dockerfile` — imagem da aplicação Go.

## Pré-requisitos

- Go 1.27+
- Git
- Docker e Docker Compose (opcional, para executar com containers)

## Como executar localmente

1. Clone o repositório:

```bash
git clone https://github.com/mariaandraddz/api_payment_go.git
cd api_payment_go
```

2. Baixe as dependências:

```bash
go mod download
```

3. Execute a aplicação:

```bash
go run .
```

A API ficará disponível em:

```text
http://localhost:8080
```

## Endpoints principais

### Saúde

- `GET /health` — verifica se a API está funcionando.

### Cartões

- `GET /card` — lista todos os cartões e suas autorizações.
- `GET /card/:id` — busca um cartão específico.
- `POST /seed` — popula os cartões em memória com dados iniciais.

### Autorizações

- `POST /authorize` — cria uma autorização de compra.

Exemplo de payload:

```json
{
  "card_id": "1",
  "amount": 100.5
}
```

### Reversão

- `POST /reversal` — cancela uma autorização já aprovada.

Exemplo:

```json
{
  "auth_id": "1",
  "amount": 100.5
}
```

### Clearing

- `POST /clearing-batch` — processa um lote de compensações.

### Settlement

- `POST /settlements-batch` — processa um lote de liquidação.
- `GET /generate-settlement-payload` — gera um payload de settlement a partir das transações já cleared.
- `GET /batch/:batch_id` — consulta o status do lote processado.

### Refund

- `POST /refund` — reembolsa um valor de uma transação já liquidada.

### Observabilidade

- `GET /metrics` — expõe métricas do Prometheus.

## Observabilidade

A aplicação coleta métricas HTTP e de negócio, incluindo:

- contadores de requisições por método e rota;
- duração de chamadas HTTP;
- contadores de autorizações aprovadas/recusadas;
- contadores de healthcheck.

Além disso, o projeto inclui configuração para Prometheus e Grafana no Docker Compose.

## Executando com Docker

Na raiz do projeto:

```bash
docker compose up --build
```

Os serviços ficam disponíveis em:

- API: `http://localhost:8080`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Observações importantes

- O projeto usa armazenamento em memória (`map`) e não possui banco de dados;
- as operações de criação de lote e processamento assíncrono são simuladas em memória;
- é um projeto focado em estudo e prototipagem de arquitetura de pagamentos.

## Licença

Este projeto está disponível para fins educacionais e de demonstração.
