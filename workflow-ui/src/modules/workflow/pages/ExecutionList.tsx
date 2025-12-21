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
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  CircularProgress,
  Alert,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Search as SearchIcon,
  Visibility as ViewIcon,
  Refresh as RefreshIcon,
} from "@mui/icons-material";
import { ExecutionState, ExecutionInfo } from "@/types/workflow";
import { workflowApi } from "@/services/client/workflowApi";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";

const ExecutionList = () => {
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [executions, setExecutions] = useState<ExecutionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const fetchExecutions = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await workflowApi.listExecutions({ projectId });
      setExecutions(response.executions || []);
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Failed to fetch executions";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchExecutions();
  }, [projectId]);

  const getStateColor = (state: string) => {
    switch (state.toUpperCase()) {
      case "SUCCESS":
      case "SUCCEEDED":
        return "success";
      case "FAILED":
        return "error";
      case "RUNNING":
        return "info";
      case "PENDING":
        return "warning";
      default:
        return "default";
    }
  };

  const normalizeState = (state: string): ExecutionState => {
    const upperState = state.toUpperCase();
    if (upperState === "SUCCESS" || upperState === "SUCCEEDED") {
      return ExecutionState.SUCCESS;
    }
    if (upperState === "FAILED") {
      return ExecutionState.FAILED;
    }
    if (upperState === "RUNNING") {
      return ExecutionState.RUNNING;
    }
    if (upperState === "PENDING") {
      return ExecutionState.PENDING;
    }
    return ExecutionState.EXECUTION_STATE_UNSPECIFIED;
  };

  const filteredExecutions = executions.filter((execution) => {
    const matchesSearch =
      execution.id?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      execution.workflowId?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      execution.workflowName?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus =
      statusFilter === "all" || normalizeState(execution.state) === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Typography variant="h4" component="h1">
          Executions
        </Typography>
        <Button
          variant="outlined"
          startIcon={<RefreshIcon />}
          onClick={fetchExecutions}
          disabled={loading}
        >
          Refresh
        </Button>
      </Box>

      <Paper sx={{ p: 2, mb: 2 }}>
        <Box sx={{ display: "flex", gap: 2, flexWrap: "wrap" }}>
          <TextField
            placeholder="Search executions..."
            size="small"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
            sx={{ flexGrow: 1, minWidth: 200 }}
          />
          <FormControl size="small" sx={{ minWidth: 150 }}>
            <InputLabel>Status</InputLabel>
            <Select
              value={statusFilter}
              label="Status"
              onChange={(e) => setStatusFilter(e.target.value)}
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value={ExecutionState.PENDING}>Pending</MenuItem>
              <MenuItem value={ExecutionState.RUNNING}>Running</MenuItem>
              <MenuItem value={ExecutionState.SUCCESS}>Success</MenuItem>
              <MenuItem value={ExecutionState.FAILED}>Failed</MenuItem>
            </Select>
          </FormControl>
        </Box>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Execution ID</TableCell>
                <TableCell>Workflow</TableCell>
                <TableCell>State</TableCell>
                <TableCell>Error</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredExecutions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                    <Typography variant="body2" color="text.secondary">
                      {searchTerm || statusFilter !== "all"
                        ? "No executions match your filters."
                        : "No executions found. Start a workflow execution to see it here."}
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                filteredExecutions.map((execution) => (
                  <TableRow key={execution.id} hover>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontFamily: "monospace" }}>
                        {execution.id}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" fontWeight="medium">
                        {execution.workflowName || execution.workflowId}
                      </Typography>
                      {execution.workflowName && (
                        <Typography variant="caption" color="text.secondary" sx={{ fontFamily: "monospace", display: "block" }}>
                          {execution.workflowId.substring(0, 8)}...
                        </Typography>
                      )}
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={execution.state}
                        color={getStateColor(execution.state) as "default" | "primary" | "secondary" | "error" | "info" | "success" | "warning"}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      {execution.error ? (
                        <Typography variant="body2" color="error" sx={{ maxWidth: 300, overflow: "hidden", textOverflow: "ellipsis" }}>
                          {execution.error}
                        </Typography>
                      ) : (
                        <Typography variant="body2" color="text.secondary">
                          -
                        </Typography>
                      )}
                    </TableCell>
                    <TableCell>
                      <IconButton
                        size="small"
                        onClick={() =>
                          navigate(`/workflows/executions/${execution.id}`)
                        }
                        title="View Details"
                      >
                        <ViewIcon />
                      </IconButton>
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

export default ExecutionList;

