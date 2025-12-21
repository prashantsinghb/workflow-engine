import { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  Grid,
  CircularProgress,
  Chip,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import { useNavigate, useParams } from "react-router-dom";
import { ExpandMore as ExpandMoreIcon } from "@mui/icons-material";
import { workflowApi } from "@/services/client/workflowApi";
import { ExecutionState, GetExecutionResponse } from "@/types/workflow";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";

const ExecutionDetails = () => {
  const { executionId } = useParams<{ executionId: string }>();
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [execution, setExecution] = useState<GetExecutionResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [polling, setPolling] = useState(false);

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

  const fetchExecution = async () => {
    try {
      const result = await workflowApi.getExecution(projectId, executionId!);
      setExecution(result);
      setLoading(false);

      // Auto-poll if still running
      if (result.state === ExecutionState.RUNNING || result.state === ExecutionState.PENDING) {
        if (!polling) {
          setPolling(true);
          const interval = setInterval(async () => {
            const updated = await workflowApi.getExecution(projectId, executionId!);
            setExecution(updated);
            if (updated.state === ExecutionState.SUCCESS || updated.state === ExecutionState.FAILED) {
              clearInterval(interval);
              setPolling(false);
            }
          }, 2000);
          return () => clearInterval(interval);
        }
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Failed to fetch execution";
      toast.error(errorMessage);
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchExecution();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [executionId]);

  if (loading) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (!execution) {
    return (
      <Box>
        <Alert severity="error">Execution not found</Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Typography variant="h4" component="h1">
          Execution Details
        </Typography>
        <Button variant="outlined" onClick={() => navigate("/workflows")}>
          Back to List
        </Button>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="text.secondary">
                  Execution ID
                </Typography>
                <Typography variant="body1">{executionId}</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="text.secondary">
                  State
                </Typography>
                <Chip
                  label={execution.state}
                  color={getStateColor(execution.state) as "default" | "primary" | "secondary" | "error" | "info" | "success" | "warning"}
                  sx={{ mt: 0.5 }}
                />
                {(execution.state === ExecutionState.RUNNING || execution.state === ExecutionState.PENDING) && (
                  <CircularProgress size={16} sx={{ ml: 1 }} />
                )}
              </Grid>
              {execution.error && (
                <Grid item xs={12}>
                  <Alert severity="error">{execution.error}</Alert>
                </Grid>
              )}
            </Grid>
          </Paper>
        </Grid>

        {execution.outputs && Object.keys(execution.outputs).length > 0 && (
          <Grid item xs={12}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Outputs
              </Typography>
              <Accordion>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                  <Typography>View Outputs</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <pre style={{ overflow: "auto", maxHeight: "400px" }}>
                    {JSON.stringify(execution.outputs, null, 2)}
                  </pre>
                </AccordionDetails>
              </Accordion>
            </Paper>
          </Grid>
        )}
      </Grid>
    </Box>
  );
};

export default ExecutionDetails;

