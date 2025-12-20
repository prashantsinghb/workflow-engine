# Setup Guide for New Service

Follow these steps to set up your new microservice from this starter framework:

## Step 1: Copy and Rename

1. Copy the entire `starter-framework` directory to your new service location
2. Rename the directory to your service name (e.g., `workflow-engine`, `notification-service`)

## Step 2: Global Find and Replace

Replace all occurrences of the following patterns:

### Pattern 1: Package/Import Paths
- `api/service` → `api/your-service-name`
- `pkg/service` → `pkg/your-service-name`
- `cmd/service` → `cmd/your-service-name`

### Pattern 2: Code Identifiers
- `service` (lowercase) → `your-service-name`
- `Service` (PascalCase) → `YourServiceName`
- `SERVICE` (UPPER_CASE) → `YOUR_SERVICE_NAME`

### Pattern 3: File/Directory Names
- Rename `api/service/` → `api/your-service-name/`
- Rename `pkg/service/` → `pkg/your-service-name/`
- Rename `cmd/service/` → `cmd/your-service-name/`
- Rename `build/service/` → `build/your-service-name/`
- Rename `helm/components/service/` → `helm/components/your-service-name/`

## Step 3: Update Proto File

1. Edit `api/your-service-name/service.proto`:
   - Update package name
   - Update service name
   - Define your API methods and messages
   - Update go_package path

2. Update `api/your-service-name/gen.go`:
   - Update the proto file name in the generate command

## Step 4: Update Go Code

1. **main.go** (`cmd/your-service-name/main.go`):
   - Update import paths
   - Update service registration
   - Update database name if needed

2. **config.go** (`pkg/your-service-name/config/config.go`):
   - Add your service-specific configuration fields
   - Add getter functions

3. **server.go** (`pkg/your-service-name/server/service.go`):
   - Implement your API handlers
   - Add dependencies (MongoDB tables, clients, etc.)

## Step 5: Update Configuration

1. **default.yml**:
   - Add your service-specific configuration

2. **helm/components/your-service-name/values.yaml**:
   - Add service-specific values

3. **helm/components/your-service-name/templates/configmap.yaml**:
   - Update configuration template

## Step 6: Update Helm Charts

1. **Chart.yaml**:
   - Update name and description

2. **deployment.yaml**:
   - Update image name
   - Update service name references

3. **service.yaml**:
   - Update service name

## Step 7: Update Dockerfile

1. **build/your-service-name/Dockerfile**:
   - Update build path: `./cmd/your-service-name/main.go`
   - Update binary name: `service` → `your-service-name`

## Step 8: Generate Code

1. Install required tools:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
   go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
   ```

2. Generate proto code:
   ```bash
   cd api/your-service-name
   go generate
   ```

## Step 9: Update go.mod

1. Update module path if needed
2. Add any additional dependencies
3. Run `go mod tidy`

## Step 10: Test Build

```bash
go build ./cmd/your-service-name/main.go
```

## Step 11: Add to Main Helm Chart (Optional)

If integrating with the main workflow-manager Helm chart:

1. Add to `helm/values.yaml`:
   ```yaml
   your-service-name:
     replicaCount:
       service: 1
     resources:
       limits:
         cpu: 300m
         memory: 2Gi
       requests:
         cpu: 50m
         memory: 400Mi
   ```

2. Add as dependency in main `Chart.yaml` if needed

## Common Patterns

### Adding MongoDB Table

1. Create `pkg/runtime/resource/table.go`:
   ```go
   package resource
   
   import (
       "go.mongodb.org/mongo-driver/bson"
       configdb "github.com/coredgeio/compass/pkg/infra/configdb"
       "github.com/coredgeio/workflow-manager/pkg/runtime"
   )
   
   type ResourceKey struct {
       Domain  string `bson:"domain,omitempty"`
       Project string `bson:"project,omitempty"`
       Name    string `bson:"name,omitempty"`
   }
   
   type ResourceEntry struct {
       Key ResourceKey `bson:"key,omitempty"`
       // Add your fields
   }
   
   type ResourceTable struct {
       *configdb.Table
   }
   
   func LocateResourceTable() (*ResourceTable, error) {
       table, err := configdb.LocateTable(runtime.WorkflowEngineDatabaseName, "resources")
       if err != nil {
           return nil, err
       }
       return &ResourceTable{Table: table}, nil
   }
   ```

2. Use in server:
   ```go
   resourceTbl, err := resource.LocateResourceTable()
   ```

### Adding WebSocket Support

See `pkg/server/websocket/` in the main workflow-manager for examples.

### Adding Background Workers

See `pkg/manager/` in the main workflow-manager for examples of background managers.

## Next Steps

1. Implement your business logic
2. Add unit tests
3. Update documentation
4. Deploy and test

