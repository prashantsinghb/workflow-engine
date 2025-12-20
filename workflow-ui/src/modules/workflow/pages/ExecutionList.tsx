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
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Search as SearchIcon,
  Visibility as ViewIcon,
  Refresh as RefreshIcon,
} from "@mui/icons-material";
import { ExecutionState } from "@/types/workflow";

interface Execution {
  id: string;
  workflowId: string;
  state: string;
  createdAt: string;
}

const ExecutionList = () => {
  const navigate = useNavigate();
  const [executions] = useState<Execution[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  useEffect(() => {
    // TODO: Implement API call to list executions
    // For now, using placeholder
  }, [statusFilter]);

  const getStateColor = (state: ExecutionState) => {
    switch (state) {
      case ExecutionState.SUCCESS:
        return "success";
      case ExecutionState.FAILED:
        return "error";
      case ExecutionState.RUNNING:
        return "info";
      case ExecutionState.PENDING:
        return "warning";
      default:
        return "default";
    }
  };

  const filteredExecutions = executions.filter((execution) => {
    const matchesSearch =
      execution.id?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      execution.workflowId?.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus =
      statusFilter === "all" || execution.state === statusFilter;
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
          onClick={() => {
            setLoading(true);
            // TODO: Refresh executions
            setTimeout(() => setLoading(false), 500);
          }}
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

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Execution ID</TableCell>
              <TableCell>Workflow ID</TableCell>
              <TableCell>State</TableCell>
              <TableCell>Created</TableCell>
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
                      {execution.id?.substring(0, 8)}...
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: "monospace" }}>
                      {execution.workflowId?.substring(0, 8)}...
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={execution.state}
                      color={getStateColor(execution.state) as "default" | "primary" | "secondary" | "error" | "info" | "success" | "warning"}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    {execution.createdAt
                      ? new Date(execution.createdAt).toLocaleString()
                      : "-"}
                  </TableCell>
                  <TableCell>
                    <IconButton
                      size="small"
                      onClick={() =>
                        navigate(`/workflows/executions/${execution.id}`)
                      }
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
    </Box>
  );
};

export default ExecutionList;

