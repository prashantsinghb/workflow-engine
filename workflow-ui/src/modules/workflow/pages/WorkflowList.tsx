import { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  TextField,
  InputAdornment,
  Card,
  CardContent,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Add as AddIcon,
  Visibility as ViewIcon,
  Search as SearchIcon,
  PlayArrow as PlayIcon,
} from "@mui/icons-material";

interface Workflow {
  workflowId: string;
  name: string;
  version: string;
}

// This is a placeholder - in a real app, you'd have an API to list workflows
// For now, we'll show a message to create workflows
const WorkflowList = () => {
  const navigate = useNavigate();
  const [workflows] = useState<Workflow[]>([]);
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    // TODO: Implement API call to list workflows
    // For now, workflows are stored locally or fetched from backend
    // Load from localStorage as a simple persistence mechanism
    const stored = localStorage.getItem("workflows");
    if (stored) {
      try {
        setWorkflows(JSON.parse(stored));
      } catch (e) {
        console.error("Failed to load workflows from storage", e);
      }
    }
  }, []);

  const filteredWorkflows = workflows.filter(
    (workflow) =>
      workflow.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      workflow.workflowId?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Typography variant="h4" component="h1">
          Workflows
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate("/workflows/create")}
        >
          Create Workflow
        </Button>
      </Box>

      {workflows.length > 0 && (
        <Paper sx={{ p: 2, mb: 2 }}>
          <TextField
            placeholder="Search workflows..."
            size="small"
            fullWidth
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
          />
        </Paper>
      )}

      {workflows.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: "center", py: 6 }}>
            <Typography variant="h6" gutterBottom>
              No Workflows Yet
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Create your first workflow to get started with the workflow engine.
            </Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => navigate("/workflows/create")}
            >
              Create Your First Workflow
            </Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Version</TableCell>
                <TableCell>Workflow ID</TableCell>
                <TableCell>Project</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredWorkflows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                    <Typography variant="body2" color="text.secondary">
                      No workflows match your search.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                filteredWorkflows.map((workflow) => (
                  <TableRow key={workflow.workflowId} hover>
                    <TableCell>
                      <Typography variant="body1" fontWeight="medium">
                        {workflow.name}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip label={workflow.version} size="small" color="primary" variant="outlined" />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontFamily: "monospace" }}>
                        {workflow.workflowId?.substring(0, 8)}...
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip label={workflow.projectId || "default"} size="small" />
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: "flex", gap: 1 }}>
                        <IconButton
                          size="small"
                          onClick={() => navigate(`/workflows/${workflow.workflowId}`)}
                          title="View Details"
                        >
                          <ViewIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => {
                            navigate(`/workflows/${workflow.workflowId}`);
                            // Scroll to execution form
                          }}
                          title="Execute"
                          color="primary"
                        >
                          <PlayIcon />
                        </IconButton>
                      </Box>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
};

export default WorkflowList;

