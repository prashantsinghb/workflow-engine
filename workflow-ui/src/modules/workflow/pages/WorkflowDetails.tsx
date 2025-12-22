import { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  Grid,
  TextField,
  CircularProgress,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Chip,
} from "@mui/material";
import { useNavigate, useParams } from "react-router-dom";
import { Formik, Form } from "formik";
import * as Yup from "yup";
import { ExpandMore as ExpandMoreIcon } from "@mui/icons-material";
import { workflowApi } from "@/services/client/workflowApi";
import { GetWorkflowResponse } from "@/types/workflow";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";
import Breadcrumbs from "@/components/atoms/Breadcrumbs";

const WorkflowDetails = () => {
  const { workflowId } = useParams<{ workflowId: string }>();
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [workflow, setWorkflow] = useState<GetWorkflowResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchWorkflow = async () => {
      if (!workflowId) {
        setError("Workflow ID is required");
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        const result = await workflowApi.getWorkflow({ projectId, workflowId });
        setWorkflow(result);
      } catch (err: unknown) {
        const errorMessage = err instanceof Error ? err.message : "Failed to load workflow";
        setError(errorMessage);
        toast.error(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchWorkflow();
  }, [workflowId, projectId]);

  const executionSchema = Yup.object({
    clientRequestId: Yup.string().required("Client Request ID is required"),
    inputs: Yup.string().test("valid-json", "Must be valid JSON", (value) => {
      if (!value) return true;
      try {
        JSON.parse(value);
        return true;
      } catch {
        return false;
      }
    }),
  });

  if (loading) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !workflow) {
    return (
      <Box sx={{ p: 3 }}>
        <Alert severity="error">{error || "Workflow not found"}</Alert>
        <Button variant="outlined" onClick={() => navigate("/workflows")} sx={{ mt: 2 }}>
          Back to List
        </Button>
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Breadcrumbs
        items={[
          { label: "Workflows", path: "/workflows" },
          { label: workflow.workflow.name || workflow.workflow.id },
        ]}
      />

      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Box>
          <Typography variant="h4" component="h1" sx={{ mb: 1 }}>
            {workflow.workflow.name || "Workflow Details"}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {workflow.workflow.id}
          </Typography>
        </Box>
        <Button variant="outlined" onClick={() => navigate("/workflows")}>
          Back to List
        </Button>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom sx={{ mb: 3 }}>
              Workflow Information
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Name
                </Typography>
                <Typography variant="body1" sx={{ fontWeight: 500 }}>
                  {workflow.workflow.name || "N/A"}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Version
                </Typography>
                <Chip label={workflow.workflow.version} size="small" color="primary" />
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Workflow ID
                </Typography>
                <Typography variant="body2" sx={{ fontFamily: "monospace", fontSize: "0.875rem" }}>
                  {workflow.workflow.id}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Project ID
                </Typography>
                <Typography variant="body2" sx={{ fontFamily: "monospace", fontSize: "0.875rem" }}>
                  {workflow.workflow.projectId}
                </Typography>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        {workflow.yaml && (
          <Grid item xs={12}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" gutterBottom>
                Workflow Definition (YAML)
              </Typography>
              <Accordion defaultExpanded>
                <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                  <Typography>View YAML Definition</Typography>
                </AccordionSummary>
                <AccordionDetails>
                  <Box
                    sx={{
                      bgcolor: "grey.50",
                      p: 2,
                      borderRadius: 1,
                      maxHeight: "600px",
                      overflow: "auto",
                    }}
                  >
                    <pre
                      style={{
                        margin: 0,
                        fontFamily: "monospace",
                        fontSize: "0.875rem",
                        whiteSpace: "pre-wrap",
                        wordBreak: "break-word",
                      }}
                    >
                      {workflow.yaml}
                    </pre>
                  </Box>
                </AccordionDetails>
              </Accordion>
            </Paper>
          </Grid>
        )}

        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom sx={{ mb: 3 }}>
              Start Execution
            </Typography>
            <Formik
              initialValues={{
                clientRequestId: `req-${Date.now()}`,
                inputs: "{}",
              }}
              validationSchema={executionSchema}
              onSubmit={async (values, { setSubmitting }) => {
                try {
                  let inputs = {};
                  if (values.inputs) {
                    inputs = JSON.parse(values.inputs);
                  }

                  const result = await workflowApi.startWorkflow({
                    projectId,
                    workflowId: workflowId!,
                    clientRequestId: values.clientRequestId,
                    inputs,
                  });

                  toast.success("Workflow execution started!");
                  navigate(`/workflows/executions/${result.executionId}`);
                } catch (error: unknown) {
                  const errorMessage = error instanceof Error ? error.message : "Failed to start workflow";
                  toast.error(errorMessage);
                } finally {
                  setSubmitting(false);
                }
              }}
            >
              {({ values, errors, touched, handleChange, isSubmitting }) => (
                <Form>
                  <Grid container spacing={3}>
                    <Grid item xs={12} md={6}>
                      <TextField
                        fullWidth
                        label="Client Request ID"
                        name="clientRequestId"
                        value={values.clientRequestId}
                        onChange={handleChange}
                        error={touched.clientRequestId && !!errors.clientRequestId}
                        helperText={touched.clientRequestId && errors.clientRequestId}
                        margin="normal"
                      />
                    </Grid>
                    <Grid item xs={12}>
                      <TextField
                        fullWidth
                        label="Inputs (JSON)"
                        name="inputs"
                        value={values.inputs}
                        onChange={handleChange}
                        error={touched.inputs && !!errors.inputs}
                        helperText={touched.inputs && errors.inputs || "Enter JSON object for workflow inputs"}
                        margin="normal"
                        multiline
                        rows={6}
                        sx={{
                          "& .MuiInputBase-root": {
                            fontFamily: "monospace",
                            fontSize: "0.875rem",
                          },
                        }}
                      />
                    </Grid>
                    <Grid item xs={12}>
                      <Button
                        variant="contained"
                        type="submit"
                        disabled={isSubmitting}
                        sx={{
                          bgcolor: "#2e7d32",
                          "&:hover": {
                            bgcolor: "#1b5e20",
                          },
                        }}
                      >
                        {isSubmitting ? (
                          <CircularProgress size={20} sx={{ color: "white" }} />
                        ) : (
                          "Start Execution"
                        )}
                      </Button>
                    </Grid>
                  </Grid>
                </Form>
              )}
            </Formik>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default WorkflowDetails;

