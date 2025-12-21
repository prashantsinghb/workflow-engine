import { useState } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  TextField,
  Grid,
  MenuItem,
  CircularProgress,
  FormControlLabel,
  Switch,
  FormHelperText,
  Chip,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Divider,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  ExpandMore as ExpandMoreIcon,
  ArrowBack as ArrowBackIcon,
  Http as HttpIcon,
  Storage as ContainerIcon,
  Info as InfoIcon,
} from "@mui/icons-material";
import { Formik, Form } from "formik";
import * as Yup from "yup";
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-github";
import { moduleApi } from "@/services/client/moduleApi";
import { toast } from "react-toastify";
import type { HttpModuleSpec, ContainerRegistryModuleSpec } from "@/types/module";

const defaultInputs = `{
  "name": "string",
  "type": "string"
}`;

const defaultOutputs = `{
  "id": "string",
  "status": "string"
}`;

const validationSchema = Yup.object({
  name: Yup.string().required("Name is required"),
  version: Yup.string().required("Version is required"),
  runtime: Yup.string().required("Runtime is required"),
  isGlobal: Yup.boolean(),
  inputs: Yup.string().test("json", "Inputs must be valid JSON", (value) => {
    if (!value) return true;
    try {
      JSON.parse(value);
      return true;
    } catch {
      return false;
    }
  }),
  outputs: Yup.string().test("json", "Outputs must be valid JSON", (value) => {
    if (!value) return true;
    try {
      JSON.parse(value);
      return true;
    } catch {
      return false;
    }
  }),
  // HTTP spec validation
  httpMethod: Yup.string().when("runtime", {
    is: "http",
    then: (schema) => schema.required("HTTP method is required"),
  }),
  httpUrl: Yup.string().when("runtime", {
    is: "http",
    then: (schema) => schema.required("HTTP URL is required").url("Must be a valid URL"),
  }),
  // Container spec validation
  containerImage: Yup.string().when("runtime", {
    is: "docker",
    then: (schema) => schema.required("Container image is required"),
  }),
});

const ModuleCreate = () => {
  const navigate = useNavigate();
  const [expandedSection, setExpandedSection] = useState<string | false>("basic");

  const handleSectionChange = (section: string) => (_event: React.SyntheticEvent, isExpanded: boolean) => {
    setExpandedSection(isExpanded ? section : false);
  };

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 4 }}>
        <Box>
          <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
            Create Module
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Define a reusable module for your workflows
          </Typography>
        </Box>
        <Button variant="outlined" onClick={() => navigate("/modules")} startIcon={<ArrowBackIcon />}>
          Cancel
        </Button>
      </Box>

      <Formik
        initialValues={{
          name: "",
          version: "v1",
          runtime: "http",
          isGlobal: false,
          inputs: defaultInputs,
          outputs: defaultOutputs,
          // HTTP spec
          httpMethod: "POST",
          httpUrl: "",
          httpHeaders: "{}",
          httpQueryParams: "{}",
          httpBodyTemplate: "{}",
          httpTimeoutMs: 30000,
          httpRetryCount: 3,
          // Auth
          authType: "none",
          authBearerToken: "",
          authApiKeyHeader: "",
          authApiKeyValue: "",
          authOAuth2TokenUrl: "",
          authOAuth2ClientId: "",
          authOAuth2ClientSecret: "",
          authOAuth2Scope: "",
          // Container spec
          containerImage: "",
          containerCommand: "",
          containerEnv: "{}",
          containerCpu: "",
          containerMemory: "",
        }}
        validationSchema={validationSchema}
        onSubmit={async (values, { setSubmitting }) => {
          try {
            let inputs = {};
            let outputs = {};
            let httpHeaders = {};
            let httpQueryParams = {};
            let httpBodyTemplate = {};
            let containerEnv = {};

            if (values.inputs) {
              try {
                inputs = JSON.parse(values.inputs);
              } catch (e) {
                toast.error("Invalid JSON in inputs");
                setSubmitting(false);
                return;
              }
            }

            if (values.outputs) {
              try {
                outputs = JSON.parse(values.outputs);
              } catch (e) {
                toast.error("Invalid JSON in outputs");
                setSubmitting(false);
                return;
              }
            }

            const projectId = values.isGlobal ? "global" : "default-project";

            let spec: HttpModuleSpec | ContainerRegistryModuleSpec | undefined;

            if (values.runtime === "http") {
              try {
                httpHeaders = JSON.parse(values.httpHeaders || "{}");
                httpQueryParams = JSON.parse(values.httpQueryParams || "{}");
                httpBodyTemplate = JSON.parse(values.httpBodyTemplate || "{}");
              } catch (e) {
                toast.error("Invalid JSON in HTTP configuration");
                setSubmitting(false);
                return;
              }

              // Build auth object
              let auth: { api_key?: { header: string; value: string }; bearer?: { token: string }; oauth2?: { token_url: string; client_id: string; client_secret?: string; scope?: string } } | undefined;
              
              if (values.authType === "bearer" && values.authBearerToken) {
                auth = { bearer: { token: values.authBearerToken } };
              } else if (values.authType === "api_key" && values.authApiKeyHeader && values.authApiKeyValue) {
                auth = { api_key: { header: values.authApiKeyHeader, value: values.authApiKeyValue } };
              } else if (values.authType === "oauth2" && values.authOAuth2TokenUrl && values.authOAuth2ClientId) {
                auth = {
                  oauth2: {
                    token_url: values.authOAuth2TokenUrl,
                    client_id: values.authOAuth2ClientId,
                    ...(values.authOAuth2ClientSecret ? { client_secret: values.authOAuth2ClientSecret } : {}),
                    ...(values.authOAuth2Scope ? { scope: values.authOAuth2Scope } : {}),
                  },
                };
              }

              spec = {
                method: values.httpMethod,
                url: values.httpUrl,
                headers: httpHeaders as Record<string, string>,
                query_params: httpQueryParams as Record<string, string>,
                body_template: httpBodyTemplate,
                ...(auth ? { auth } : {}),
                timeout_ms: values.httpTimeoutMs,
                retry_count: values.httpRetryCount,
              };
            } else if (values.runtime === "docker") {
              try {
                containerEnv = JSON.parse(values.containerEnv || "{}");
              } catch (e) {
                toast.error("Invalid JSON in container environment variables");
                setSubmitting(false);
                return;
              }

              const command = values.containerCommand
                ? values.containerCommand.split("\n").filter((c) => c.trim())
                : undefined;

              spec = {
                image: values.containerImage,
                command,
                env: containerEnv as Record<string, string>,
                cpu: values.containerCpu || undefined,
                memory: values.containerMemory || undefined,
              };
            }

            await moduleApi.registerModule({
              projectId,
              module: {
                name: values.name,
                version: values.version,
                runtime: values.runtime,
                ...(values.runtime === "http" ? { http: spec as HttpModuleSpec } : { container_registry: spec as ContainerRegistryModuleSpec }),
                inputs,
                outputs,
              },
            });

            toast.success("Module created successfully!");
            navigate(`/modules/${values.name}/versions/${values.version}`);
          } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : "Failed to create module";
            toast.error(errorMessage);
          } finally {
            setSubmitting(false);
          }
        }}
      >
        {({ values, errors, touched, handleChange, setFieldValue, isSubmitting }) => (
          <Form>
            <Grid container spacing={3}>
              {/* Basic Information */}
              <Grid item xs={12}>
                <Accordion expanded={expandedSection === "basic"} onChange={handleSectionChange("basic")} elevation={2}>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                      <InfoIcon color="primary" />
                      <Typography variant="h6" sx={{ fontWeight: 500 }}>
                        Basic Information
                      </Typography>
                    </Box>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Grid container spacing={3}>
                      <Grid item xs={12} md={6}>
                        <TextField
                          fullWidth
                          label="Module Name"
                          name="name"
                          value={values.name}
                          onChange={handleChange}
                          error={touched.name && !!errors.name}
                          helperText={touched.name && errors.name || "e.g., compute.create"}
                          placeholder="e.g., compute.create"
                          variant="outlined"
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
                          helperText={touched.version && errors.version || "e.g., v1"}
                          placeholder="e.g., v1"
                          variant="outlined"
                        />
                      </Grid>
                      <Grid item xs={12} md={3}>
                        <Box sx={{ display: "flex", flexDirection: "column", justifyContent: "center", height: "100%" }}>
                          <FormControlLabel
                            control={
                              <Switch
                                checked={values.isGlobal}
                                onChange={(e) => setFieldValue("isGlobal", e.target.checked)}
                                color="primary"
                              />
                            }
                            label="Global Module"
                          />
                          <FormHelperText sx={{ ml: 0, mt: -0.5 }}>
                            {values.isGlobal ? "Available to all projects" : "Project-specific"}
                          </FormHelperText>
                        </Box>
                      </Grid>
                    </Grid>
                  </AccordionDetails>
                </Accordion>
              </Grid>

              {/* Runtime Selection */}
              <Grid item xs={12}>
                <Accordion expanded={expandedSection === "runtime"} onChange={handleSectionChange("runtime")} elevation={2}>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                      {values.runtime === "http" ? <HttpIcon color="primary" /> : <ContainerIcon color="primary" />}
                      <Typography variant="h6" sx={{ fontWeight: 500 }}>
                        Runtime Configuration
                      </Typography>
                      <Chip label={values.runtime === "http" ? "HTTP" : "Container"} size="small" color="primary" sx={{ ml: 1 }} />
                    </Box>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Grid container spacing={3}>
                      <Grid item xs={12} md={4}>
                        <TextField
                          fullWidth
                          select
                          label="Runtime Type"
                          name="runtime"
                          value={values.runtime}
                          onChange={handleChange}
                          error={touched.runtime && !!errors.runtime}
                          helperText={touched.runtime && errors.runtime}
                          variant="outlined"
                        >
                          <MenuItem value="http">
                            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                              <HttpIcon fontSize="small" />
                              <span>HTTP</span>
                            </Box>
                          </MenuItem>
                          <MenuItem value="docker">
                            <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                              <ContainerIcon fontSize="small" />
                              <span>Container Image</span>
                            </Box>
                          </MenuItem>
                        </TextField>
                      </Grid>

                      {/* HTTP Configuration */}
                      {values.runtime === "http" && (
                        <>
                          <Grid item xs={12} md={3}>
                            <TextField
                              fullWidth
                              select
                              label="HTTP Method"
                              name="httpMethod"
                              value={values.httpMethod}
                              onChange={handleChange}
                              variant="outlined"
                            >
                              <MenuItem value="GET">GET</MenuItem>
                              <MenuItem value="POST">POST</MenuItem>
                              <MenuItem value="PUT">PUT</MenuItem>
                              <MenuItem value="DELETE">DELETE</MenuItem>
                              <MenuItem value="PATCH">PATCH</MenuItem>
                            </TextField>
                          </Grid>
                          <Grid item xs={12} md={5}>
                            <TextField
                              fullWidth
                              label="URL"
                              name="httpUrl"
                              value={values.httpUrl}
                              onChange={handleChange}
                              error={touched.httpUrl && !!errors.httpUrl}
                              helperText={touched.httpUrl && errors.httpUrl || "e.g., https://api.example.com/execute"}
                              placeholder="https://api.example.com/execute"
                              variant="outlined"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <Box>
                              <Typography variant="subtitle2" gutterBottom>
                                Headers (JSON)
                              </Typography>
                              <AceEditor
                                mode="json"
                                theme="github"
                                value={values.httpHeaders}
                                onChange={(value) => setFieldValue("httpHeaders", value)}
                                width="100%"
                                height="150px"
                                fontSize={13}
                                setOptions={{ showLineNumbers: true, tabSize: 2 }}
                              />
                            </Box>
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <Box>
                              <Typography variant="subtitle2" gutterBottom>
                                Query Parameters (JSON)
                              </Typography>
                              <AceEditor
                                mode="json"
                                theme="github"
                                value={values.httpQueryParams}
                                onChange={(value) => setFieldValue("httpQueryParams", value)}
                                width="100%"
                                height="150px"
                                fontSize={13}
                                setOptions={{ showLineNumbers: true, tabSize: 2 }}
                              />
                            </Box>
                          </Grid>
                          <Grid item xs={12}>
                            <Box>
                              <Typography variant="subtitle2" gutterBottom>
                                Body Template (JSON)
                              </Typography>
                              <AceEditor
                                mode="json"
                                theme="github"
                                value={values.httpBodyTemplate}
                                onChange={(value) => setFieldValue("httpBodyTemplate", value)}
                                width="100%"
                                height="200px"
                                fontSize={13}
                                setOptions={{ showLineNumbers: true, tabSize: 2 }}
                              />
                            </Box>
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              fullWidth
                              label="Timeout (ms)"
                              name="httpTimeoutMs"
                              type="number"
                              value={values.httpTimeoutMs}
                              onChange={handleChange}
                              variant="outlined"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              fullWidth
                              label="Retry Count"
                              name="httpRetryCount"
                              type="number"
                              value={values.httpRetryCount}
                              onChange={handleChange}
                              variant="outlined"
                            />
                          </Grid>
                          {/* Authentication */}
                          <Grid item xs={12}>
                            <Divider sx={{ my: 2 }} />
                            <Typography variant="subtitle1" gutterBottom sx={{ fontWeight: 600, mb: 2 }}>
                              Authentication
                            </Typography>
                          </Grid>
                          <Grid item xs={12} md={4}>
                            <TextField
                              fullWidth
                              select
                              label="Auth Type"
                              name="authType"
                              value={values.authType}
                              onChange={handleChange}
                              variant="outlined"
                            >
                              <MenuItem value="none">None</MenuItem>
                              <MenuItem value="bearer">Bearer Token</MenuItem>
                              <MenuItem value="api_key">API Key</MenuItem>
                              <MenuItem value="oauth2">OAuth2</MenuItem>
                            </TextField>
                          </Grid>
                          {values.authType === "bearer" && (
                            <Grid item xs={12} md={8}>
                              <TextField
                                fullWidth
                                label="Bearer Token"
                                name="authBearerToken"
                                type="password"
                                value={values.authBearerToken}
                                onChange={handleChange}
                                placeholder="Enter bearer token"
                                variant="outlined"
                              />
                            </Grid>
                          )}
                          {values.authType === "api_key" && (
                            <>
                              <Grid item xs={12} md={6}>
                                <TextField
                                  fullWidth
                                  label="Header Name"
                                  name="authApiKeyHeader"
                                  value={values.authApiKeyHeader}
                                  onChange={handleChange}
                                  placeholder="e.g., X-API-Key"
                                  variant="outlined"
                                />
                              </Grid>
                              <Grid item xs={12} md={6}>
                                <TextField
                                  fullWidth
                                  label="API Key Value"
                                  name="authApiKeyValue"
                                  type="password"
                                  value={values.authApiKeyValue}
                                  onChange={handleChange}
                                  placeholder="Enter API key"
                                  variant="outlined"
                                />
                              </Grid>
                            </>
                          )}
                          {values.authType === "oauth2" && (
                            <>
                              <Grid item xs={12} md={6}>
                                <TextField
                                  fullWidth
                                  label="Token URL"
                                  name="authOAuth2TokenUrl"
                                  value={values.authOAuth2TokenUrl}
                                  onChange={handleChange}
                                  placeholder="https://oauth.example.com/token"
                                  variant="outlined"
                                />
                              </Grid>
                              <Grid item xs={12} md={6}>
                                <TextField
                                  fullWidth
                                  label="Client ID"
                                  name="authOAuth2ClientId"
                                  value={values.authOAuth2ClientId}
                                  onChange={handleChange}
                                  placeholder="Enter client ID"
                                  variant="outlined"
                                />
                              </Grid>
                              <Grid item xs={12} md={6}>
                                <TextField
                                  fullWidth
                                  label="Client Secret"
                                  name="authOAuth2ClientSecret"
                                  type="password"
                                  value={values.authOAuth2ClientSecret}
                                  onChange={handleChange}
                                  placeholder="Enter client secret (optional)"
                                  variant="outlined"
                                />
                              </Grid>
                              <Grid item xs={12} md={6}>
                                <TextField
                                  fullWidth
                                  label="Scope"
                                  name="authOAuth2Scope"
                                  value={values.authOAuth2Scope}
                                  onChange={handleChange}
                                  placeholder="e.g., read write (optional)"
                                  variant="outlined"
                                />
                              </Grid>
                            </>
                          )}
                        </>
                      )}

                      {/* Container Configuration */}
                      {values.runtime === "docker" && (
                        <>
                          <Grid item xs={12}>
                            <TextField
                              fullWidth
                              label="Container Image"
                              name="containerImage"
                              value={values.containerImage}
                              onChange={handleChange}
                              error={touched.containerImage && !!errors.containerImage}
                              helperText={touched.containerImage && errors.containerImage || "e.g., my-image:latest"}
                              placeholder="my-image:latest"
                              variant="outlined"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              fullWidth
                              label="Command (one per line)"
                              name="containerCommand"
                              value={values.containerCommand}
                              onChange={handleChange}
                              helperText="Enter each command on a new line"
                              multiline
                              rows={4}
                              variant="outlined"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <Box>
                              <Typography variant="subtitle2" gutterBottom>
                                Environment Variables (JSON)
                              </Typography>
                              <AceEditor
                                mode="json"
                                theme="github"
                                value={values.containerEnv}
                                onChange={(value) => setFieldValue("containerEnv", value)}
                                width="100%"
                                height="150px"
                                fontSize={13}
                                setOptions={{ showLineNumbers: true, tabSize: 2 }}
                              />
                            </Box>
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              fullWidth
                              label="CPU (e.g., 100m, 1)"
                              name="containerCpu"
                              value={values.containerCpu}
                              onChange={handleChange}
                              placeholder="100m"
                              variant="outlined"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              fullWidth
                              label="Memory (e.g., 128Mi, 1Gi)"
                              name="containerMemory"
                              value={values.containerMemory}
                              onChange={handleChange}
                              placeholder="128Mi"
                              variant="outlined"
                            />
                          </Grid>
                        </>
                      )}
                    </Grid>
                  </AccordionDetails>
                </Accordion>
              </Grid>

              {/* Schema Definition */}
              <Grid item xs={12}>
                <Accordion expanded={expandedSection === "schema"} onChange={handleSectionChange("schema")} elevation={2}>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                      <InfoIcon color="primary" />
                      <Typography variant="h6" sx={{ fontWeight: 500 }}>
                        Schema Definition
                      </Typography>
                    </Box>
                  </AccordionSummary>
                  <AccordionDetails>
                    <Grid container spacing={3}>
                      <Grid item xs={12} md={6}>
                        <Box>
                          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                            Inputs (JSON Schema)
                          </Typography>
                          <AceEditor
                            mode="json"
                            theme="github"
                            value={values.inputs}
                            onChange={(value) => setFieldValue("inputs", value)}
                            width="100%"
                            height="300px"
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
                          {touched.inputs && errors.inputs && (
                            <Typography variant="caption" color="error" sx={{ mt: 1, display: "block" }}>
                              {errors.inputs}
                            </Typography>
                          )}
                        </Box>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Box>
                          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                            Outputs (JSON Schema)
                          </Typography>
                          <AceEditor
                            mode="json"
                            theme="github"
                            value={values.outputs}
                            onChange={(value) => setFieldValue("outputs", value)}
                            width="100%"
                            height="300px"
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
                          {touched.outputs && errors.outputs && (
                            <Typography variant="caption" color="error" sx={{ mt: 1, display: "block" }}>
                              {errors.outputs}
                            </Typography>
                          )}
                        </Box>
                      </Grid>
                    </Grid>
                  </AccordionDetails>
                </Accordion>
              </Grid>

              {/* Action Buttons */}
              <Grid item xs={12}>
                <Paper sx={{ p: 2, bgcolor: "background.default" }}>
                  <Box sx={{ display: "flex", justifyContent: "flex-end", gap: 2 }}>
                    <Button
                      variant="outlined"
                      onClick={() => navigate("/modules")}
                      size="large"
                    >
                      Cancel
                    </Button>
                    <Button
                      variant="contained"
                      type="submit"
                      disabled={isSubmitting}
                      size="large"
                      sx={{ minWidth: 150 }}
                    >
                      {isSubmitting ? <CircularProgress size={20} /> : "Create Module"}
                    </Button>
                  </Box>
                </Paper>
              </Grid>
            </Grid>
          </Form>
        )}
      </Formik>
    </Box>
  );
};

export default ModuleCreate;
