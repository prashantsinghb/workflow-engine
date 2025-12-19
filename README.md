# Workflow Engine

A distributed workflow orchestration engine built with Go, gRPC, and Temporal. This engine enables you to define, validate, and execute complex workflows as Directed Acyclic Graphs (DAGs) with support for parallel execution, dependency management, and fault tolerance.

## Features

- **DAG-based Workflows**: Define workflows as directed acyclic graphs with node dependencies
- **Workflow Validation**: Validate workflow definitions for cycles, missing dependencies, and structural integrity
- **Parallel Execution**: Execute independent nodes in parallel for optimal performance
- **Temporal Integration**: Built-in support for Temporal workflows for durable, fault-tolerant execution
- **REST & gRPC APIs**: Dual API support with gRPC for high-performance and REST for ease of use
- **OpenAPI Documentation**: Auto-generated Swagger/OpenAPI documentation
- **Executor Registry**: Pluggable executor system for custom node implementations
- **Project-based Organization**: Multi-tenant support with project isolation

## Architecture

```
┌─────────────────┐
│   REST Client   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────┐
│  gRPC Gateway   │────▶│  gRPC Server │
└─────────────────┘     └──────┬───────┘
                                │
                                ▼
                        ┌───────────────┐
                        │ Workflow      │
                        │ Server        │
                        └───────┬───────┘
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
        ┌───────────┐   ┌───────────┐   ┌───────────┐
        │  Parser   │   │    DAG    │   │ Execution │
        │           │   │  Engine   │   │  Engine   │
        └───────────┘   └───────────┘   └───────────┘
                                │
                                ▼
                        ┌───────────────┐
                        │   Temporal     │
                        │   Workflows    │
                        └───────────────┘
```

## Project Structure

```
workflow-engine/
├── api/
│   ├── openapi/              # OpenAPI/Swagger documentation
│   └── service/              # Protocol Buffer definitions
│       ├── service.proto     # gRPC service definitions
│       └── gen.go            # Code generation script
├── cmd/
│   └── service/              # Main application entry point
│       └── main.go
├── pkg/
│   ├── config/               # Configuration management
│   ├── server/                # gRPC server implementation
│   └── workflow/
│       ├── api/              # Workflow API models
│       ├── dag/              # DAG operations (graph, validate, plan, ready)
│       ├── execution/        # Execution engine
│       ├── executor/         # Executor registry and implementations
│       ├── parser/           # YAML workflow parser
│       ├── registry/         # Workflow registry
│       └── temporal/        # Temporal workflow integration
├── build/
│   └── service/
│       └── Dockerfile       # Container build file
├── helm/
│   └── components/
│       └── service/        # Helm chart for Kubernetes deployment
├── third_party/            # Third-party proto definitions
└── default.yml             # Default configuration
```

## Installation

### Prerequisites

- Go 1.24+ 
- Protocol Buffers compiler (`protoc`)
- Protocol Buffer plugins:
  - `protoc-gen-go`
  - `protoc-gen-go-grpc`
  - `protoc-gen-grpc-gateway`
  - `protoc-gen-openapiv2`

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/prashantsinghb/workflow-engine.git
   cd workflow-engine
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Generate protocol buffer code**:
   ```bash
   go generate ./api/service
   ```

4. **Build the service**:
   ```bash
   go build ./cmd/service
   ```

5. **Run the service**:
   ```bash
   ./service
   ```

The service will start:
- gRPC server on `:50051`
- REST API gateway on `:8080`
- Swagger UI at `http://localhost:8080/swagger/`

## Usage

### Workflow Definition Format

Workflows are defined in YAML format:

```yaml
nodes:
  step1:
    uses: compute.create
    with:
      name: "instance-1"
      type: "t2.micro"
  
  step2:
    uses: compute.create
    depends_on:
      - step1
    with:
      name: "instance-2"
      type: "t2.micro"
  
  step3:
    uses: network.configure
    depends_on:
      - step1
      - step2
    with:
      vpc: "default"
```

### API Examples

#### 1. Validate a Workflow

```bash
curl -X POST http://localhost:8080/v1/projects/my-project/workflows:validate \
  -H "Content-Type: application/json" \
  -d '{
    "workflow": {
      "name": "my-workflow",
      "version": "1.0.0",
      "yaml": "nodes:\n  step1:\n    uses: compute.create"
    }
  }'
```

#### 2. Register a Workflow

```bash
curl -X POST http://localhost:8080/v1/projects/my-project/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "workflow": {
      "name": "my-workflow",
      "version": "1.0.0",
      "yaml": "nodes:\n  step1:\n    uses: compute.create"
    }
  }'
```

Response:
```json
{
  "workflow_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### 3. Start Workflow Execution

```bash
curl -X POST http://localhost:8080/v1/projects/my-project/executions \
  -H "Content-Type: application/json" \
  -d '{
    "workflow_id": "550e8400-e29b-41d4-a716-446655440000",
    "inputs": {
      "region": "us-east-1"
    }
  }'
```

Response:
```json
{
  "execution_id": "660e8400-e29b-41d4-a716-446655440001"
}
```

#### 4. Get Execution Status

```bash
curl http://localhost:8080/v1/projects/my-project/executions/660e8400-e29b-41d4-a716-446655440001
```

Response:
```json
{
  "state": "SUCCEEDED",
  "outputs": {
    "step1": {
      "instance_id": "i-1234567890abcdef0"
    }
  }
}
```

### gRPC API

The service also exposes a gRPC API. See `api/service/service.proto` for the complete API definition.

## Development

### Code Generation

After modifying `service.proto`, regenerate the code:

```bash
go generate ./api/service
```

### Running Tests

```bash
go test ./...
```

### Building Docker Image

```bash
docker build -t workflow-engine:latest -f build/service/Dockerfile .
```

### Kubernetes Deployment

Deploy using Helm:

```bash
helm install workflow-engine ./helm/components/service
```

## DAG Operations

The engine provides several DAG operations:

- **Build**: Constructs a DAG from workflow definition
- **Validate**: Checks for cycles and missing dependencies
- **Topological Sort**: Orders nodes for execution
- **Ready Nodes**: Identifies nodes ready to execute (all dependencies satisfied)

## Executors

Executors are pluggable components that execute individual workflow nodes. Register custom executors:

```go
executor.Register("my-executor", &MyExecutor{})
```

Built-in executors:
- `noop`: No-operation executor for testing

## Temporal Integration

The engine supports Temporal workflows for durable, fault-tolerant execution. Temporal workflows provide:
- Automatic retries
- Workflow history
- Long-running workflow support
- Activity timeouts and retries

## Configuration

Edit `default.yml` for service configuration:

```yaml
# Service configuration
server:
  grpc_port: 50051
  http_port: 8080
```

## API Documentation

Interactive API documentation is available at:
- Swagger UI: `http://localhost:8080/swagger/`
- OpenAPI JSON: `http://localhost:8080/swagger/apidocs.swagger.json`

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[Add your license here]

## Support

For issues and questions, please open an issue on GitHub.
