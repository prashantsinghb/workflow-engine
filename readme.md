# Workflow Engine

A distributed workflow orchestration platform built with Go, React, and Temporal. This platform enables you to define, validate, and execute complex workflows as Directed Acyclic Graphs (DAGs) with support for parallel execution, dependency management, fault tolerance, and a modern web interface.

## 🚀 Features

### Core Capabilities
- **DAG-based Workflows**: Define workflows as directed acyclic graphs with node dependencies
- **Workflow Validation**: Validate workflow definitions for cycles, missing dependencies, and structural integrity
- **Parallel Execution**: Execute independent nodes in parallel for optimal performance
- **Temporal Integration**: Built-in support for Temporal workflows for durable, fault-tolerant execution
- **Module System**: Reusable modules (HTTP, Container Registry) that can be composed into workflows
- **Template Rendering**: Dynamic template rendering with support for nested inputs and step outputs
- **Authentication Support**: HTTP modules support Bearer, API Key, and OAuth2 authentication

### User Interface
- **Modern Web UI**: React-based interface with Material-UI components
- **Workflow Management**: Create, validate, and manage workflow definitions
- **Execution Monitoring**: Real-time execution status tracking with auto-polling
- **Dashboard**: Overview with workflow and execution statistics
- **Project-based Organization**: Multi-tenant support with project isolation
- **YAML Editor**: Built-in YAML editor with syntax highlighting

### API & Integration
- **REST & gRPC APIs**: Dual API support with gRPC for high-performance and REST for ease of use
- **OpenAPI Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Executor Registry**: Pluggable executor system for custom node implementations
- **Project Context**: Centralized project management across the UI

## 📋 Table of Contents

- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Quick Start](#quick-start)
- [Components](#components)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Web Browser                          │
│                    (React UI - Port 3000)                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP/REST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Workflow Engine API                      │
│  ┌──────────────┐         ┌──────────────┐                │
│  │ gRPC Gateway │────────▶│  gRPC Server │                │
│  │  (Port 8081) │         │  (Port 50051)│                │
│  └──────────────┘         └──────┬───────┘                │
└───────────────────────────────────┼─────────────────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    ▼               ▼               ▼
        ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
        │   Parser     │   │     DAG      │   │  Execution   │
        │              │   │   Engine     │   │   Engine     │
        └──────────────┘   └──────────────┘   └──────┬───────┘
                                                       │
                    ┌─────────────────────────────────┼──────────┐
                    ▼                                 ▼          ▼
        ┌──────────────────┐              ┌──────────────────┐
        │   PostgreSQL     │              │    Temporal      │
        │   (Workflows,    │              │    (Orchestration)│
        │   Executions,    │              │                  │
        │   Modules)       │              │                  │
        └──────────────────┘              └──────────────────┘
```

## 📁 Project Structure

```
.
├── workflow-engine/          # Backend Go service
│   ├── api/                  # Protocol Buffer definitions
│   │   ├── openapi/         # OpenAPI/Swagger documentation
│   │   └── service/          # gRPC service definitions
│   ├── cmd/
│   │   ├── service/          # Main API server
│   │   └── worker/          # Temporal worker
│   ├── pkg/
│   │   ├── config/          # Configuration management
│   │   ├── execution/       # Execution storage and models
│   │   ├── module/          # Module registry and management
│   │   ├── server/          # gRPC server implementation
│   │   └── workflow/
│   │       ├── dag/         # DAG operations (graph, validate, plan, ready)
│   │       ├── execution/  # Execution engine
│   │       ├── executor/    # Executor registry and implementations
│   │       │   ├── http_executor.go    # HTTP module executor
│   │       │   ├── template.go         # Template rendering
│   │       │   └── http_auth.go        # HTTP authentication
│   │       ├── parser/      # YAML workflow parser
│   │       ├── registry/   # Workflow registry
│   │       └── temporal/   # Temporal workflow integration
│   ├── migration/           # Database migrations
│   ├── build/               # Docker build files
│   └── helm/                # Kubernetes Helm charts
│
├── workflow-ui/             # Frontend React application
│   ├── src/
│   │   ├── components/      # Reusable UI components
│   │   │   ├── atoms/       # Atomic components (Logo, etc.)
│   │   │   └── layouts/    # Layout components (MainLayout)
│   │   ├── modules/         # Feature modules
│   │   │   ├── workflow/   # Workflow management
│   │   │   └── module/     # Module management
│   │   ├── services/        # API services
│   │   ├── contexts/        # React contexts (ProjectContext)
│   │   ├── types/          # TypeScript type definitions
│   │   └── routes/         # Application routes
│   └── public/             # Static assets
│
└── docker-compose/          # Temporal server docker-compose files
    └── deployment/          # Monitoring and observability stack
        ├── grafana/        # Grafana dashboards
        ├── prometheus/     # Prometheus configuration
        └── loki/           # Loki logging
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+** (for backend)
- **Node.js 18+** (for frontend)
- **PostgreSQL** (for data storage)
- **Temporal Server** (for workflow orchestration)
- **Protocol Buffers compiler** (`protoc`) and plugins

### 1. Start Infrastructure

Start PostgreSQL and Temporal Server using docker-compose:

```bash
cd docker-compose
docker-compose -f docker-compose-postgres.yml up -d
```

This will start:
- PostgreSQL on port 5432
- Temporal Server on port 7233
- Temporal Web UI on port 8080

### 2. Setup Database

Run database migrations:

```bash
cd workflow-engine
# Set your database connection string
export DATABASE_URL="postgres://user:password@localhost:5432/workflow_engine?sslmode=disable"

# Run migrations (adjust based on your migration tool)
psql $DATABASE_URL < migration/001_create_executions.sql
psql $DATABASE_URL < migration/002_create_modules.sql
psql $DATABASE_URL < migration/003_create_module_http_specs.sql
psql $DATABASE_URL < migration/004_create_module_container_registry_specs.sql
psql $DATABASE_URL < migration/005_create_workflows.sql
psql $DATABASE_URL < migration/006_create_execution_nodes.sql
psql $DATABASE_URL < migration/007_create_execution_events.sql
psql $DATABASE_URL < migration/008_create_execution_artifacts.sql
```

### 3. Start Backend

```bash
cd workflow-engine

# Install dependencies
go mod download

# Generate protocol buffer code
make generate

# Build and run the service
go run ./cmd/service/main.go
```

The backend will start:
- gRPC server on `:50051`
- REST API gateway on `:8081`
- Swagger UI at `http://localhost:8081/swagger/`

### 4. Start Temporal Worker

In a separate terminal:

```bash
cd workflow-engine
go run ./cmd/worker/main.go
```

### 5. Start Frontend

```bash
cd workflow-ui

# Install dependencies
npm install

# Create .env file
cat > .env << EOF
VITE_PORT=3000
VITE_API_BASE_URL=http://localhost:8081
EOF

# Start development server
npm start
```

The UI will be available at `http://localhost:3000`

## 🧩 Components

### Backend (workflow-engine)

The backend is a Go-based microservice that provides:

- **Workflow Management**: Register, validate, and retrieve workflows
- **Execution Engine**: Start and monitor workflow executions
- **Module Registry**: Manage reusable modules (HTTP, Container Registry)
- **DAG Processing**: Build, validate, and execute DAG-based workflows
- **Temporal Integration**: Orchestrate workflows using Temporal

**Key Technologies:**
- Go 1.24
- gRPC & gRPC-Gateway
- Protocol Buffers
- Temporal SDK
- PostgreSQL
- FastTemplate (for template rendering)

### Frontend (workflow-ui)

The frontend is a React-based single-page application that provides:

- **Dashboard**: Overview of workflows, executions, and success rates
- **Workflow Management**: Create, list, and view workflows
- **Execution Monitoring**: Track execution status and view outputs
- **Module Management**: Create and manage reusable modules
- **Project Context**: Switch between projects

**Key Technologies:**
- React 18
- TypeScript
- Material-UI (MUI)
- Vite
- React Router
- Formik & Yup
- Axios

### Infrastructure

- **PostgreSQL**: Stores workflows, executions, and modules
- **Temporal**: Orchestrates workflow execution with durability and fault tolerance
- **Docker Compose**: Local development infrastructure setup

## 💻 Development

### Backend Development

#### Code Generation

After modifying `.proto` files:

```bash
cd workflow-engine
make generate
```

#### Running Tests

```bash
go test ./...
```

#### Building Docker Image

```bash
docker build -t workflow-engine:latest -f workflow-engine/build/service/Dockerfile workflow-engine/
```

### Frontend Development

#### Available Scripts

```bash
npm start      # Start development server
npm run build  # Build for production
npm run preview # Preview production build
npm run lint   # Run ESLint
```

#### Environment Variables

Create a `.env` file in `workflow-ui/`:

```env
VITE_PORT=3000
VITE_API_BASE_URL=http://localhost:8081
```

## 📚 API Documentation

### REST API

The REST API is available at `http://localhost:8081` with interactive Swagger documentation at `http://localhost:8081/swagger/`

#### Key Endpoints

**Workflows:**
- `POST /v1/projects/{projectId}/workflows:validate` - Validate workflow
- `POST /v1/projects/{projectId}/workflows` - Register workflow
- `GET /v1/projects/{projectId}/workflows` - List workflows
- `GET /v1/projects/{projectId}/workflows/{workflowId}` - Get workflow

**Executions:**
- `POST /v1/projects/{projectId}/executions` - Start execution
- `GET /v1/projects/{projectId}/executions` - List executions
- `GET /v1/projects/{projectId}/executions/{executionId}` - Get execution status

**Modules:**
- `POST /v1/projects/{projectId}/modules` - Register module
- `GET /v1/projects/{projectId}/modules` - List modules
- `GET /v1/projects/{projectId}/modules/{name}/versions/{version}` - Get module

**Dashboard:**
- `GET /v1/projects/{projectId}/dashboard/stats` - Get dashboard statistics

### gRPC API

The gRPC API is available on port `50051`. See `workflow-engine/api/service/service.proto` for the complete API definition.

## 📝 Workflow Definition Format

Workflows are defined in YAML format:

```yaml
nodes:
  step1:
    uses: http.postman-echo
    with:
      url: "https://postman-echo.com/post"
      headers:
        custom-header: "123"
      body_template:
        message: "{{inputs.message}}"
  
  step2:
    uses: http.postman-echo
    depends_on:
      - step1
    with:
      url: "https://postman-echo.com/get"
      query_params:
        param1: "{{steps.step1.data.message}}"
```

### Module Types

**HTTP Modules:**
- Support GET, POST, PUT, DELETE methods
- Template rendering for body, headers, and query parameters
- Authentication: Bearer, API Key, OAuth2
- Retry logic and timeouts

**Container Registry Modules:**
- Execute containerized workloads
- Support for Docker images

## 🎯 Features in Detail

### Template Rendering

The engine supports dynamic template rendering using `{{ }}` syntax:

- `{{inputs.field}}` - Access workflow inputs
- `{{steps.stepName.field}}` - Access previous step outputs
- Nested access: `{{inputs.nested.field}}`

### Execution States

- `PENDING` - Execution created but not started
- `RUNNING` - Execution in progress
- `SUCCEEDED` - Execution completed successfully
- `FAILED` - Execution failed

### Project Isolation

All resources (workflows, executions, modules) are scoped to projects, enabling multi-tenant support.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go code style guidelines
- Write tests for new features
- Update documentation for API changes
- Ensure all tests pass before submitting PR

## 📄 License

[Add your license here]

## 🔗 Related Documentation

- [Backend README](workflow-engine/README.md)
- [Frontend README](workflow-ui/README.md)
- [Backend Setup Guide](workflow-engine/SETUP.md)
- [Frontend Setup Guide](workflow-ui/SETUP.md)
- [Temporal Documentation](https://docs.temporal.io/)

## 🆘 Support

For issues and questions:
- Open an issue on GitHub
- Check existing documentation in component READMEs
- Review API documentation at `/swagger/`

---

**Built with ❤️ using Go, React, and Temporal**

