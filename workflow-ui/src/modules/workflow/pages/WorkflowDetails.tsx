import { useState, useEffect } from "react";
import {
  Box,
  Container,
  Typography,
  Button,
  Paper,
  Grid,
  TextField,
  CircularProgress,
  Alert,
} from "@mui/material";
import { useNavigate, useParams } from "react-router-dom";
import { Formik, Form } from "formik";
import * as Yup from "yup";
import { workflowApi } from "@/services/client/workflowApi";
import { GetWorkflowResponse } from "@/types/workflow";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";

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
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4, textAlign: "center" }}>
        <CircularProgress />
      </Container>
    );
  }

  if (error || !workflow) {
    return (
      <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
        <Alert severity="error">{error || "Workflow not found"}</Alert>
        <Button variant="outlined" onClick={() => navigate("/workflows")} sx={{ mt: 2 }}>
          Back to List
        </Button>
      </Container>
    );
  }

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
        <Typography variant="h4" component="h1">
          Workflow Details
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
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Workflow Name
                </Typography>
                <Typography variant="h6">{workflow.workflow.name}</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Version
                </Typography>
                <Typography variant="body1">{workflow.workflow.version}</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Workflow ID
                </Typography>
                <Typography variant="body2" sx={{ fontFamily: "monospace" }}>
                  {workflow.workflow.id}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Project ID
                </Typography>
                <Typography variant="body1">{workflow.workflow.projectId}</Typography>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        <Grid item xs={12}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
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
                  <Grid container spacing={2}>
                    <Grid item xs={12}>
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
                        helperText={touched.inputs && errors.inputs}
                        margin="normal"
                        multiline
                        rows={4}
                      />
                    </Grid>
                    <Grid item xs={12}>
                      <Button
                        variant="contained"
                        type="submit"
                        disabled={isSubmitting}
                      >
                        {isSubmitting ? <CircularProgress size={20} /> : "Start Execution"}
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

