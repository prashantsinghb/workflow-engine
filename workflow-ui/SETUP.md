# Setup Guide

## Quick Start

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Create environment file**:
   ```bash
   # Create .env file with:
   VITE_PORT=3000
   VITE_API_BASE_URL=http://localhost:8080
   ```

3. **Start the development server**:
   ```bash
   npm start
   ```

4. **Open your browser**:
   Navigate to `http://localhost:3000`

## Prerequisites

- Node.js 18+ 
- npm or yarn
- Workflow Engine backend running on port 8080 (or update `VITE_API_BASE_URL`)

## Project Structure

```
workflow-ui/
├── src/
│   ├── components/          # Reusable UI components
│   ├── modules/             # Feature modules
│   │   └── workflow/       # Workflow management module
│   ├── services/            # API services
│   ├── themes/              # Material-UI theme
│   ├── types/               # TypeScript type definitions
│   └── routes/              # Application routes
├── public/                  # Static assets
└── package.json
```

## Features

- ✅ Workflow list view
- ✅ Create workflow with YAML editor
- ✅ Validate workflow before registration
- ✅ Start workflow execution
- ✅ Monitor execution status with auto-polling
- ✅ View execution outputs

## API Endpoints Used

- `POST /v1/projects/{projectId}/workflows:validate` - Validate workflow
- `POST /v1/projects/{projectId}/workflows` - Register workflow
- `POST /v1/projects/{projectId}/executions` - Start execution
- `GET /v1/projects/{projectId}/executions/{executionId}` - Get execution status

## Troubleshooting

### Port already in use
Change `VITE_PORT` in `.env` to a different port.

### Cannot connect to backend
1. Verify backend is running
2. Check `VITE_API_BASE_URL` in `.env`
3. Check CORS settings on backend

### Module not found errors
Run `npm install` again to ensure all dependencies are installed.

