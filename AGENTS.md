# AGENTS.md

This file provides guidance to Qoder (qoder.com) when working with code in this repository.

## Build & Development Commands

### Local Development
```bash
make up                    # Start all services (LocalStack, Postgres, Redis, subgraphs)
make down                  # Stop all services
make seed                  # Seed database with initial data
make deploy                # Deploy to LocalStack using SAM
```

### Database Migrations
```bash
go run scripts/migrate/migrate.go    # Run migrations for both write and read models
```

### Testing & Invoking Lambdas
```bash
make invoke-cmd func=interventionCreate payload='{"patient_id":"123","screening_id":"456","items":[]}'
make invoke-worker func=userEventWorker payload='{"event":"user_created"}'
```

### Building Individual Lambdas
```bash
cd cmd/<lambda-name> && GOOS=linux GOARCH=amd64 go build -o main
sam build                            # Build all lambdas defined in template.yaml
```

## Architecture Overview

This is a **CQRS event-driven Lambda architecture** for a healthcare application with multi-tenant support. The system separates write-side commands from read-side queries and uses Apollo Federation for GraphQL.

### Key Patterns

1. **CQRS (Command Query Responsibility Segregation)**
   - Write Model: `postgres_write` (port 5434) - handles commands
   - Read Model: `postgres_read` (port 5433) - handles queries
   - Event streaming via Kinesis to sync read/write models

2. **Multi-tenancy**
   - All domain models have `tenant_id` for data isolation
   - Tenant context extracted from JWT tokens or headers (`X-Tenant-ID`)

3. **Authentication**
   - AWS Cognito with custom auth flows (OTP-based)
   - Custom Cognito triggers: PreSignUp, PostConfirmation, DefineAuthChallenge, CreateAuthChallenge, VerifyAuthChallenge
   - JWT token validation for API Gateway

### Directory Structure

```
/cmd                    # Write-side Lambda handlers (commands)
  ├─ authInvite/       # Auth flow: invite user
  ├─ authLogin/        # Auth flow: login
  ├─ authRegister/     # Auth flow: register
  ├─ interventionCreate/   # Create interventions
  ├─ interventionUpdate/   # Update interventions
  └─ cognitoXXX/       # Cognito trigger handlers

/query                  # Read-side Lambda handlers (queries) - NOT YET IMPLEMENTED

/workers                # Asynchronous event processors (SQS/EventBridge consumers)
  └─ userEventWorker/  # Processes user events from SQS

/apps                   # GraphQL subgraphs (Apollo Federation)
  ├─ gateway/          # Apollo Router/Gateway config
  ├─ subgraph-auth/    # Auth & Users domain
  ├─ subgraph-intervention/  # Interventions domain
  ├─ subgraph-patient/       # Patient domain (stub)
  ├─ subgraph-screening/     # Screening domain (stub)
  └─ subgraph-navigator/     # Navigator workbench (stub)

/internal               # ALL BUSINESS LOGIC
  ├─ domain/           # Domain models (Patient, Intervention, User, etc.)
  ├─ service/          # Business rules & workflows
  ├─ repository/       # Database access layer (GORM-based)
  ├─ db/               # DB connections & migrations
  ├─ events/           # Event definitions & publishing
  ├─ auth/             # JWT & Cognito utilities
  ├─ validator/        # Input validation
  ├─ logger/           # HIPAA-safe logging
  └─ errors/           # Typed error definitions

/scripts/migrate        # SQL migration files (.up.sql, .down.sql)
```

### How to Add a New Lambda

#### For Command Lambdas (Write Operations)
1. Create directory under `/cmd/<operation-name>/`
2. Create `main.go` with handler that:
   - Accepts `events.APIGatewayProxyRequest`
   - Extracts tenant_id and user_id from headers/context
   - Calls service layer in `/internal/service`
   - Returns `events.APIGatewayProxyResponse`
3. Add function definition to `template.yaml` under Resources
4. Build with `GOOS=linux GOARCH=amd64 go build -o main`

#### For Query Lambdas (Read Operations)
1. Create directory under `/query/<operation-name>/`
2. Query the read model database (`postgres_read`)
3. Follow same pattern as command lambdas

#### For Workers (Background Jobs)
1. Create directory under `/workers/<worker-name>/`
2. Process events from SQS or EventBridge
3. Call service layer to update read models or trigger side effects

### How to Add a New Subgraph

1. Create directory under `/apps/subgraph-<domain>/`
2. Initialize with gqlgen: `go run github.com/99designs/gqlgen init`
3. Define schema in `schema.graphqls` with Federation directives
4. Implement resolvers in `graph/schema.resolvers.go`
5. Add Dockerfile for containerization
6. Add service to `docker-compose.yml`
7. Update gateway configuration in `/apps/gateway/supergraph.graphqls`

### Domain Models

All domain models are in `/internal/domain/models.go`:
- **Intervention**: Core entity with type (financial_counselor, social_work, etc.), status (pending, in_progress, completed, cancelled), and multi-tenant support
- **Patient**: Patient records
- **Screening**: Screening assessments
- **User**: System users with roles

### Service Layer Conventions

Services in `/internal/service/` follow this pattern:
```go
type XService struct {
    repo *repository.XRepository
}

func (s *XService) Operation(ctx context.Context, tenantID string, ...) error {
    // Business logic here
    // Call repository for persistence
}
```

### Repository Layer Conventions

Repositories in `/internal/repository/` use GORM:
```go
type XRepository struct {
    db *gorm.DB
}

func (r *XRepository) Create(ctx context.Context, entity *domain.X) error {
    return r.db.WithContext(ctx).Create(entity).Error
}
```

Always filter by `tenant_id` for multi-tenant queries.

### Environment Variables

Copy `.env.example` to `.env` and configure:

**Database (CQRS)**:
- `WRITE_DB_HOST`, `WRITE_DB_PORT`, `WRITE_DB_USER`, `WRITE_DB_PASSWORD`, `WRITE_DB_NAME`: Write model DB
- `READ_DB_HOST`, `READ_DB_PORT`, `READ_DB_USER`, `READ_DB_PASSWORD`, `READ_DB_NAME`: Read model DB

**Redis Cache**:
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`: Redis connection

**AWS / LocalStack**:
- `LOCALSTACK_URL`: LocalStack endpoint (default: http://localhost:4566)
- `AWS_REGION`: AWS region (default: us-east-1)
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`: AWS credentials (use "test" for local)

**Authentication**:
- `COGNITO_USER_POOL_ID`, `COGNITO_CLIENT_ID`: Cognito configuration
- `JWT_SECRET`: JWT signing key

**Note**: When running in Docker, database hosts should be service names (e.g., `postgres_write`). When running locally, use `localhost`.

### Infrastructure

- **LocalStack**: AWS service emulation for local development (port 4566)
- **PostgreSQL**: Two instances (write: 5434, read: 5433)
- **Redis**: Caching layer (port 6380)
- **SAM (Serverless Application Model)**: Deployment framework

### Important Notes

1. **HIPAA Compliance**: Use HIPAA-safe logging (no PHI in logs)
2. **Tenant Isolation**: Always filter by tenant_id in queries
3. **Event Sourcing**: Commands emit events to Kinesis; workers project to read model
4. **Apollo Federation**: Subgraphs must use `@key`, `@external`, `@requires` directives
5. **Go Version**: 1.24.2 (see go.mod)
6. **Lambda Runtime**: go1.x architecture x86_64
