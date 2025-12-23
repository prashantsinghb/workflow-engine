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
  Grid,
  Card,
  CardContent,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Search as SearchIcon,
  MoreVert as MoreVertIcon,
  Refresh as RefreshIcon,
  ViewList as ListViewIcon,
  ViewModule as GridViewIcon,
  Visibility as ViewIcon,
} from "@mui/icons-material";
import { ExecutionState, ExecutionInfo } from "@/types/workflow";
import { workflowApi } from "@/services/client/workflowApi";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";
import Breadcrumbs from "@/components/atoms/Breadcrumbs";

const ExecutionList = () => {
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [executions, setExecutions] = useState<ExecutionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [viewMode, setViewMode] = useState<"list" | "grid">("list");

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
    if (upperState === "SUCCESS") {
      return ExecutionState.SUCCESS;
    }
    if (upperState === "SUCCEEDED") {
      return ExecutionState.SUCCEEDED;
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
    const normalizedState = normalizeState(execution.state);
    const matchesStatus =
      statusFilter === "all" ||
      normalizedState === statusFilter ||
      (statusFilter === ExecutionState.SUCCESS && normalizedState === ExecutionState.SUCCEEDED) ||
      (statusFilter === ExecutionState.SUCCEEDED && normalizedState === ExecutionState.SUCCESS);
    return matchesSearch && matchesStatus;
  });

  return (
    <Box sx={{ p: 3 }}>
      <Breadcrumbs items={[{ label: "Dashboard", path: "/" }, { label: "Executions" }]} />
      
      <Box sx={{ mb: 3 }}>
        <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", mb: 1 }}>
          <Box>
            <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
              Executions
            </Typography>
            <Typography variant="body2" color="text.secondary">
              View and manage workflow execution history
            </Typography>
          </Box>
        </Box>
      </Box>

      <Box sx={{ display: "flex", gap: 2, mb: 2, alignItems: "center", flexWrap: "wrap" }}>
        <TextField
          placeholder="Search"
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
          sx={{ flexGrow: 1, maxWidth: 400 }}
        />
        <FormControl size="small" sx={{ minWidth: 150 }}>
          <InputLabel>State</InputLabel>
          <Select
            value={statusFilter}
            label="State"
            onChange={(e) => setStatusFilter(e.target.value)}
          >
            <MenuItem value="all">All</MenuItem>
            <MenuItem value={ExecutionState.PENDING}>Pending</MenuItem>
            <MenuItem value={ExecutionState.RUNNING}>Running</MenuItem>
            <MenuItem value={ExecutionState.SUCCESS}>Success</MenuItem>
            <MenuItem value={ExecutionState.SUCCEEDED}>Succeeded</MenuItem>
            <MenuItem value={ExecutionState.FAILED}>Failed</MenuItem>
          </Select>
        </FormControl>
        <IconButton
          onClick={() => setViewMode("list")}
          color={viewMode === "list" ? "primary" : "default"}
        >
          <ListViewIcon />
        </IconButton>
        <IconButton
          onClick={() => setViewMode("grid")}
          color={viewMode === "grid" ? "primary" : "default"}
        >
          <GridViewIcon />
        </IconButton>
        <IconButton onClick={fetchExecutions} disabled={loading}>
          <RefreshIcon />
        </IconButton>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {loading ? (
        <Box sx={{ display: "flex", justifyContent: "center", py: 4 }}>
          <CircularProgress />
        </Box>
      ) : viewMode === "list" ? (
        <TableContainer component={Paper} sx={{ boxShadow: "none", border: "1px solid #e0e0e0" }}>
          <Table>
            <TableHead>
              <TableRow sx={{ backgroundColor: "#fafafa" }}>
                <TableCell sx={{ fontWeight: 600 }}>Execution ID</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Workflow</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>State</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Error</TableCell>
                <TableCell sx={{ fontWeight: 600, width: 100 }} align="right">Actions</TableCell>
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
                      <Typography variant="body2" sx={{ fontFamily: "monospace", color: "text.secondary" }}>
                        {execution.id}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Typography variant="body1" fontWeight={500}>
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
                    <TableCell align="right">
                      <IconButton
                        size="small"
                        onClick={() =>
                          navigate(`/workflows/executions/${execution.id}`)
                        }
                        title="View Details"
                        color="primary"
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
      ) : (
        <Grid container spacing={3}>
          {filteredExecutions.length === 0 ? (
            <Grid item xs={12}>
              <Box sx={{ textAlign: "center", py: 4 }}>
                <Typography variant="body2" color="text.secondary">
                  {searchTerm || statusFilter !== "all"
                    ? "No executions match your filters."
                    : "No executions found. Start a workflow execution to see it here."}
                </Typography>
              </Box>
            </Grid>
          ) : (
            filteredExecutions.map((execution) => (
              <Grid item xs={12} sm={6} md={4} key={execution.id}>
                <Card sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
                  <CardContent sx={{ flexGrow: 1 }}>
                    <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", mb: 2 }}>
                      <Typography variant="h6" component="div" sx={{ fontWeight: 600, flexGrow: 1 }}>
                        {execution.workflowName || execution.workflowId}
                      </Typography>
                      <Chip
                        label={execution.state}
                        color={getStateColor(execution.state) as "default" | "primary" | "secondary" | "error" | "info" | "success" | "warning"}
                        size="small"
                      />
                    </Box>
                    <Box sx={{ mb: 2 }}>
                      <Typography variant="caption" color="text.secondary" sx={{ display: "block", mb: 0.5 }}>
                        Execution ID
                      </Typography>
                      <Typography variant="body2" sx={{ fontFamily: "monospace", color: "text.secondary" }}>
                        {execution.id}
                      </Typography>
                    </Box>
                    {execution.workflowName && (
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="caption" color="text.secondary" sx={{ display: "block", mb: 0.5 }}>
                          Workflow ID
                        </Typography>
                        <Typography variant="body2" sx={{ fontFamily: "monospace", color: "text.secondary" }}>
                          {execution.workflowId.substring(0, 16)}...
                        </Typography>
                      </Box>
                    )}
                    {execution.error && (
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="caption" color="text.secondary" sx={{ display: "block", mb: 0.5 }}>
                          Error
                        </Typography>
                        <Typography variant="body2" color="error" sx={{ wordBreak: "break-word" }}>
                          {execution.error}
                        </Typography>
                      </Box>
                    )}
                    <Button
                      variant="contained"
                      size="small"
                      startIcon={<ViewIcon />}
                      onClick={() => navigate(`/workflows/executions/${execution.id}`)}
                      fullWidth
                      sx={{
                        backgroundColor: "#2e7d32",
                        "&:hover": {
                          backgroundColor: "#1b5e20",
                        },
                        mt: "auto",
                      }}
                    >
                      View Details
                    </Button>
                  </CardContent>
                </Card>
              </Grid>
            ))
          )}
        </Grid>
      )}
    </Box>
  );
};

export default ExecutionList;

