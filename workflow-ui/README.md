# Workflow Engine UI

A modern React-based user interface for the Workflow Engine backend. This UI provides an intuitive interface for creating, validating, and executing workflows.

## Features

- **Workflow Management**: Create and manage workflow definitions
- **YAML Editor**: Built-in YAML editor with syntax highlighting for workflow definitions
- **Workflow Validation**: Validate workflows before registration
- **Execution Management**: Start and monitor workflow executions
- **Real-time Status**: Auto-polling for execution status updates
- **Modern UI**: Built with Material-UI for a clean, responsive interface

## Tech Stack

- **React 18** - UI framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Material-UI (MUI)** - Component library
- **React Router** - Routing
- **Formik & Yup** - Form handling and validation
- **Axios** - HTTP client
- **React Ace** - Code editor for YAML

## Prerequisites

- Node.js 18+ and npm/yarn
- Workflow Engine backend running (default: http://localhost:8080)

## Installation

1. **Install dependencies**:
   ```bash
   cd workflow-ui
   npm install
   ```

2. **Configure environment**:
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` and set:
   ```
   VITE_PORT=3000
   VITE_API_BASE_URL=http://localhost:8080
   ```

3. **Start development server**:
   ```bash
   npm start
   ```

   The UI will be available at `http://localhost:3000`

## Usage

### Creating a Workflow

1. Navigate to the Workflows page
2. Click "Create Workflow"
3. Fill in:
   - Workflow Name
   - Version
   - Project ID
   - YAML definition
4. Click "Validate" to check the workflow before creating
5. Click "Create Workflow" to register it

### Example Workflow YAML

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

### Executing a Workflow

1. Navigate to a workflow's details page
2. Fill in:
   - Client Request ID (auto-generated)
   - Inputs (optional JSON)
3. Click "Start Execution"
4. View execution status and outputs on the execution details page

## Project Structure

```
workflow-ui/
├── src/
│   ├── components/          # Reusable components
│   │   └── layouts/         # Layout components
│   ├── modules/             # Feature modules
│   │   └── workflow/        # Workflow module
│   │       ├── pages/      # Page components
│   │       └── routes.tsx  # Route definitions
│   ├── services/            # API services
│   │   └── client/         # API client
│   ├── themes/              # Theme configuration
│   ├── types/               # TypeScript types
│   ├── routes/              # Main routing
│   ├── App.tsx              # Root component
│   └── index.tsx            # Entry point
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## API Integration

The UI integrates with the Workflow Engine backend API:

- `POST /v1/projects/{projectId}/workflows:validate` - Validate workflow
- `POST /v1/projects/{projectId}/workflows` - Register workflow
- `POST /v1/projects/{projectId}/executions` - Start execution
- `GET /v1/projects/{projectId}/executions/{executionId}` - Get execution status

## Development

### Available Scripts

- `npm start` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

### Building for Production

```bash
npm run build
```

The built files will be in the `dist` directory, ready for deployment.

## Configuration

### Environment Variables

- `VITE_PORT` - Development server port (default: 3000)
- `VITE_API_BASE_URL` - Backend API URL (default: http://localhost:8080)

## Troubleshooting

### CORS Issues

If you encounter CORS errors, ensure the backend is configured to allow requests from the UI origin.

### API Connection

Verify that:
1. The backend is running
2. `VITE_API_BASE_URL` in `.env` matches the backend URL
3. The backend is accessible from your browser

## License

[Add your license here]

