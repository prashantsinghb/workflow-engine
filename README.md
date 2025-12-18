# Starter Framework for Microservices

This is a starter template for creating new microservices following the workflow-manager framework patterns.

## Structure

```
starter-framework/
├── api/
│   └── service/              # Protocol Buffer definitions
│       ├── service.proto     # Main service API definition
│       └── gen.go            # Code generation script
├── cmd/
│   └── service/              # Main application entry point
│       └── main.go
├── pkg/
│   ├── config/               # Configuration management
│   │   └── config.go
│   └── server/                # gRPC server implementations
│       └── service.go
├── build/
│   └── service/
│       └── Dockerfile        # Container build file
├── helm/
│   └── components/
│       └── service/           # Helm chart for deployment
│           ├── Chart.yaml
│           ├── values.yaml
│           └── templates/
│               ├── configmap.yaml
│               ├── deployment.yaml
│               └── service.yaml
└── default.yml                # Default configuration file
```

## Quick Start

1. **Copy the starter-framework directory** to your new service location
2. **Rename the service** - Replace all occurrences of:
   - `service` → your service name
   - `Service` → YourService (PascalCase)
   - `SERVICE` → YOUR_SERVICE (UPPER_CASE)
3. **Update the proto file** (`api/service/service.proto`) with your API definitions
4. **Implement your server logic** in `pkg/server/service.go`
5. **Update configuration** in `pkg/config/config.go` and `default.yml`
6. **Customize Helm charts** in `helm/components/service/`
7. **Generate proto code**: Run `go generate ./api/service/`
8. **Build and deploy**

## Configuration

Edit `default.yml` and `pkg/config/config.go` to add your service-specific configuration.

## API Development

1. Define your API in `api/service/service.proto`
2. Run `go generate ./api/service/` to generate Go code
3. Implement handlers in `pkg/server/service.go`

## Deployment

Use the Helm chart in `helm/components/service/` to deploy to Kubernetes.

## Dependencies

Make sure to add any additional dependencies to `go.mod` as needed.

