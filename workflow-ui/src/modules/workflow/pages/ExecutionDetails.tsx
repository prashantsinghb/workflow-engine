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
import { ExecutionState, GetExecutionResponse, ExecutionTimeline } from "@/types/workflow";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";
import { ExecutionTimeline as ExecutionTimelineComponent } from "../components/ExecutionTimeline";
import Breadcrumbs from "@/components/atoms/Breadcrumbs";

const ExecutionDetails = () => {
  const { executionId } = useParams<{ executionId: string }>();
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [execution, setExecution] = useState<GetExecutionResponse | null>(null);
  const [timeline, setTimeline] = useState<ExecutionTimeline | null>(null);
  const [loading, setLoading] = useState(true);
  const [timelineLoading, setTimelineLoading] = useState(false);
  const [polling, setPolling] = useState(false);

  const getStateColor = (state: ExecutionState | string) => {
    const stateStr = String(state).toUpperCase();
    switch (stateStr) {
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

  const fetchTimeline = async () => {
    try {
      setTimelineLoading(true);
      const result = await workflowApi.getExecutionTimeline(projectId, executionId!);
      setTimeline(result);
    } catch (error: unknown) {
      // Timeline is optional, don't show error if it fails
      console.error("Failed to fetch timeline:", error);
    } finally {
      setTimelineLoading(false);
    }
  };

  const fetchExecution = async () => {
    try {
      const result = await workflowApi.getExecution(projectId, executionId!);
      setExecution(result);
      setLoading(false);

      // Fetch timeline
      fetchTimeline();

            // Auto-poll if still running
      const stateStr = String(result.state);
      if (stateStr === "RUNNING" || stateStr === "PENDING") {
        if (!polling) {
          setPolling(true);
          const interval = setInterval(async () => {
            const updated = await workflowApi.getExecution(projectId, executionId!);
            setExecution(updated);
            // Refresh timeline on each poll
            fetchTimeline();
            const updatedStateStr = String(updated.state);
            if (updatedStateStr === "SUCCESS" || updatedStateStr === "SUCCEEDED" || updatedStateStr === "FAILED") {
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
    <Box sx={{ p: 3 }}>
      <Breadcrumbs
        items={[
          { label: "Workflows", path: "/workflows" },
          { label: "Executions", path: "/workflows/executions" },
          { label: executionId?.substring(0, 8) || "Execution" },
        ]}
      />

      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Box>
          <Typography variant="h4" component="h1" sx={{ mb: 1 }}>
            Execution Details
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ fontFamily: "monospace" }}>
            {executionId}
          </Typography>
        </Box>
        <Button variant="outlined" onClick={() => navigate("/workflows/executions")}>
          Back to List
        </Button>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom sx={{ mb: 3 }}>
              Execution Information
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Execution ID
                </Typography>
                <Typography variant="body2" sx={{ fontFamily: "monospace", fontSize: "0.875rem" }}>
                  {executionId}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  State
                </Typography>
                <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                  <Chip
                    label={String(execution.state)}
                    color={getStateColor(execution.state) as "default" | "primary" | "secondary" | "error" | "info" | "success" | "warning"}
                    size="small"
                  />
                  {(() => {
                    const stateStr = String(execution.state);
                    return (stateStr === "RUNNING" || stateStr === "PENDING") && (
                      <CircularProgress size={16} />
                    );
                  })()}
                </Box>
              </Grid>
              {execution.error && (
                <Grid item xs={12}>
                  <Alert severity="error">{execution.error}</Alert>
                </Grid>
              )}
            </Grid>
          </Paper>
        </Grid>

        {timeline && (
          <Grid item xs={12}>
            {timelineLoading ? (
              <Paper sx={{ p: 3, textAlign: "center" }}>
                <CircularProgress size={24} />
              </Paper>
            ) : (
              <ExecutionTimelineComponent timeline={timeline} />
            )}
          </Grid>
        )}

        {execution.inputs && Object.keys(execution.inputs).length > 0 && (
          <Grid item xs={12}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Inputs
              </Typography>
              <Accordion>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                  <Typography>View Inputs</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <pre style={{ overflow: "auto", maxHeight: "400px" }}>
                    {JSON.stringify(execution.inputs, null, 2)}
                  </pre>
                </AccordionDetails>
              </Accordion>
            </Paper>
          </Grid>
        )}

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

