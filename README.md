# YouTube Clone — Spec-Driven Development

Clone do YouTube construído com Spec-Driven Development (SDD). O arquivo `openapi.yaml` é a fonte da verdade — toda funcionalidade começa nele.

## Arquitetura

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Frontend   │────▶│   Backend   │────▶│  LocalStack  │
│   Angular    │     │  Go / Echo  │     │  S3 + Dynamo │
│  :4200       │     │  :3000      │     │  :4566       │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Stack

| Camada     | Tecnologia                                      |
|------------|------------------------------------------------|
| Frontend   | Angular 18 (Standalone Components)             |
| Backend    | Go 1.22 + Echo v4 + oapi-codegen v2            |
| Infra      | Docker Compose + Floci (S3 + DynamoDB local)    |
| API Spec   | OpenAPI 3.0.3                                   |

## Pré-requisitos

- Go >= 1.22
- Node.js >= 20
- Docker + Docker Compose

## Como Rodar

### 1. Infraestrutura (LocalStack)

```bash
docker compose up -d
```

Cria container Floci com S3 e DynamoDB em `http://localhost:4566`.

### 2. Backend

```bash
cd backend
go run .
```

Servidor inicia em `http://localhost:3000`. Bucket `youtube-clone-videos` e tabela `Videos` são criados automaticamente.

### 3. Frontend

```bash
cd frontend
npm install
npm start
```

Angular em `http://localhost:4200`. Proxy configurado para encaminhar `/s3/*` ao LocalStack.

## API

### Endpoints

| Método | Rota                   | Descrição                     |
|--------|------------------------|-------------------------------|
| POST   | `/videos/upload`       | Gera URL presigned para S3    |
| POST   | `/videos`              | Registra metadados do vídeo   |
| GET    | `/videos`              | Lista vídeos (paginado)       |
| GET    | `/videos/{id}`         | Busca vídeo por ID            |
| POST   | `/videos/{id}/like`    | Incrementa curtidas           |

### Exemplos via cURL

**Upload:**
```bash
curl -X POST http://localhost:3000/videos/upload \
  -H 'Content-Type: application/json' \
  -d '{"fileName":"video.mp4","fileType":"video/mp4"}'
```

**Registrar vídeo:**
```bash
curl -X POST http://localhost:3000/videos \
  -H 'Content-Type: application/json' \
  -d '{
    "id":"d290f1ee-6c54-4b01-90e6-d701748f0851",
    "titulo":"Meu vídeo",
    "url_s3":"https://bucket.s3.amazonaws.com/abc.mp4",
    "autor":"kaike"
  }'
```

**Listar vídeos:**
```bash
curl http://localhost:3000/videos?limit=10
```

**Curtir vídeo:**
```bash
curl -X POST http://localhost:3000/videos/d290f1ee-6c54-4b01-90e6-d701748f0851/like
```

## Schema Video

```yaml
Video:
  id: UUID
  titulo: string
  descricao: string (nullable)
  url_s3: string (uri)
  autor: string
  thumbnailUrl: string (nullable)
  viewCount: integer
  likeCount: integer
  createdAt: datetime
  updatedAt: datetime
```

## Ciclo SDD

Toda nova funcionalidade segue este fluxo:

1. **Atualizar `openapi.yaml`** — Definir rotas, schemas e responses
2. **Gerar código** — Executar geradores do backend e frontend
3. **Implementar lógica** — Preencher handlers Go e componentes Angular

### Comandos de geração

**Backend (Go):**
```bash
cd backend
$(go env GOPATH)/bin/oapi-codegen --config oapi-codegen.cfg.yaml ../openapi.yaml
```

**Frontend (Angular):**
```bash
cd frontend
npx openapi-generator-cli generate -i ../openapi.yaml -g typescript-angular -o src/app/core/api
```

## Estrutura do Projeto

```
clone-youtube-sdd/
├── openapi.yaml                    # Fonte da verdade (OpenAPI 3.0.3)
├── docker-compose.yml              # Floci (S3 + DynamoDB local)
├── backend/
│   ├── main.go                     # Handlers Echo + rotas
│   ├── aws.go                      # Clients S3 e DynamoDB
│   ├── api/
│   │   └── gen.go                  # Código gerado pelo oapi-codegen
│   ├── oapi-codegen.cfg.yaml       # Config do gerador Go
│   ├── go.mod / go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/app/
│   │   ├── features/
│   │   │   ├── video-upload/       # Componente de upload
│   │   │   └── video-list/         # Componente de listagem + like
│   │   ├── core/api/               # Código gerado pelo openapi-generator
│   │   ├── app.component.ts
│   │   └── app.config.ts           # provideApi + provideHttpClient
│   ├── proxy.conf.json             # Proxy /s3 → LocalStack
│   └── angular.json
└── README.md
```

## Licença

MIT
