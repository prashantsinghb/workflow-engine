import { useState } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  TextField,
  Grid,
  Alert,
  CircularProgress,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import { Formik, Form } from "formik";
import * as Yup from "yup";
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/mode-yaml";
import "ace-builds/src-noconflict/theme-github";
import { workflowApi } from "@/services/client/workflowApi";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";
import Breadcrumbs from "@/components/atoms/Breadcrumbs";

const defaultYaml = `nodes:
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
`;

const validationSchema = Yup.object({
  name: Yup.string().required("Name is required"),
  version: Yup.string().required("Version is required"),
  yaml: Yup.string().required("YAML definition is required"),
  projectId: Yup.string().required("Project ID is required"),
});

const WorkflowCreate = () => {
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [validating, setValidating] = useState(false);
  const [validationResult, setValidationResult] = useState<{ valid: boolean; errors: string[] } | null>(null);

  const handleValidate = async (yaml: string, projectId: string) => {
    setValidating(true);
    setValidationResult(null);
    try {
      const result = await workflowApi.validateWorkflow({
        projectId,
        workflow: {
          name: "temp",
          version: "1.0.0",
          yaml,
        },
      });
      setValidationResult(result);
      if (result.valid) {
        toast.success("Workflow is valid!");
      } else {
        toast.error("Workflow validation failed");
      }
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Validation failed";
      toast.error(errorMessage);
    } finally {
      setValidating(false);
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Breadcrumbs items={[{ label: "Dashboard", path: "/" }, { label: "Workflows", path: "/workflows" }, { label: "Create Workflow" }]} />
      <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
        Create Workflow
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
        Define a new workflow using YAML format
      </Typography>

      <Formik
        enableReinitialize
        initialValues={{
          name: "",
          version: "1.0.0",
          yaml: defaultYaml,
          projectId: projectId,
        }}
        validationSchema={validationSchema}
        onSubmit={async (values, { setSubmitting }) => {
          try {
            // First validate
            const validation = await workflowApi.validateWorkflow({
              projectId: values.projectId,
              workflow: {
                name: values.name,
                version: values.version,
                yaml: values.yaml,
              },
            });

            if (!validation.valid) {
              toast.error("Workflow validation failed. Please fix the errors.");
              setValidationResult(validation);
              setSubmitting(false);
              return;
            }

            // Then register
            const result = await workflowApi.registerWorkflow({
              projectId: values.projectId,
              workflow: {
                name: values.name,
                version: values.version,
                yaml: values.yaml,
              },
            });

            toast.success("Workflow created successfully!");
            navigate(`/workflows/${result.workflowId}`);
          } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : "Failed to create workflow";
            toast.error(errorMessage);
          } finally {
            setSubmitting(false);
          }
        }}
      >
        {({ values, errors, touched, handleChange, setFieldValue, isSubmitting }) => (
          <Form>
            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <TextField
                  fullWidth
                  label="Workflow Name"
                  name="name"
                  value={values.name}
                  onChange={handleChange}
                  error={touched.name && !!errors.name}
                  helperText={touched.name && errors.name}
                  margin="normal"
                />
              </Grid>
              <Grid item xs={12} md={3}>
                <TextField
                  fullWidth
                  label="Version"
                  name="version"
                  value={values.version}
                  onChange={handleChange}
                  error={touched.version && !!errors.version}
                  helperText={touched.version && errors.version}
                  margin="normal"
                />
              </Grid>
              <Grid item xs={12} md={3}>
                <TextField
                  fullWidth
                  label="Project ID"
                  name="projectId"
                  value={values.projectId}
                  onChange={handleChange}
                  error={touched.projectId && !!errors.projectId}
                  helperText={touched.projectId && errors.projectId}
                  margin="normal"
                  disabled
                />
              </Grid>
              <Grid item xs={12}>
                <Paper sx={{ p: 2 }}>
                  <Box sx={{ mb: 2, display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <Typography variant="h6">Workflow YAML Definition</Typography>
                    <Button
                      variant="outlined"
                      onClick={() => handleValidate(values.yaml, values.projectId)}
                      disabled={validating}
                    >
                      {validating ? <CircularProgress size={20} /> : "Validate"}
                    </Button>
                  </Box>
                  <AceEditor
                    mode="yaml"
                    theme="github"
                    value={values.yaml}
                    onChange={(value) => setFieldValue("yaml", value)}
                    width="100%"
                    height="400px"
                    fontSize={14}
                    showPrintMargin={true}
                    showGutter={true}
                    highlightActiveLine={true}
                    setOptions={{
                      enableBasicAutocompletion: true,
                      enableLiveAutocompletion: true,
                      enableSnippets: true,
                      showLineNumbers: true,
                      tabSize: 2,
                    }}
                  />
                </Paper>
              </Grid>
              {validationResult && (
                <Grid item xs={12}>
                  {validationResult.valid ? (
                    <Alert severity="success">Workflow is valid!</Alert>
                  ) : (
                    <Alert severity="error">
                      <Typography variant="subtitle2" gutterBottom>
                        Validation Errors:
                      </Typography>
                      <ul>
                        {validationResult.errors.map((error, index) => (
                          <li key={index}>{error}</li>
                        ))}
                      </ul>
                    </Alert>
                  )}
                </Grid>
              )}
              <Grid item xs={12}>
                <Box sx={{ display: "flex", gap: 2 }}>
                  <Button
                    variant="contained"
                    type="submit"
                    disabled={isSubmitting}
                    sx={{
                      backgroundColor: "#2e7d32",
                      color: "#ffffff",
                      textTransform: "none",
                      "&:hover": {
                        backgroundColor: "#1b5e20",
                      },
                    }}
                  >
                    {isSubmitting ? <CircularProgress size={20} /> : "Create Workflow"}
                  </Button>
                  <Button variant="outlined" onClick={() => navigate("/workflows")}>
                    Cancel
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </Form>
        )}
      </Formik>
    </Box>
  );
};

export default WorkflowCreate;

