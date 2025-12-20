import { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  Grid,
  Chip,
  CircularProgress,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import {
  ExpandMore as ExpandMoreIcon,
  Http as HttpIcon,
  Storage as ContainerIcon,
  ArrowBack as ArrowBackIcon,
} from "@mui/icons-material";
import { useNavigate, useParams } from "react-router-dom";
import AceEditor from "react-ace";
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-github";
import { moduleApi } from "@/services/client/moduleApi";
import { toast } from "react-toastify";
import type { Module } from "@/types/module";

const ModuleDetails = () => {
  const { name, version } = useParams<{ name: string; version: string }>();
  const navigate = useNavigate();
  const [module, setModule] = useState<Module | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const projectId = "default-project"; // TODO: Get from context or route

  useEffect(() => {
    if (name && version) {
      loadModule();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [name, version]);

  const loadModule = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await moduleApi.getModule({
        projectId,
        name: name!,
        version: version!,
      });
      setModule(response.module);
    } catch (error: unknown) {
      console.error("Failed to load module", error);
      const errorMessage = error instanceof Error ? error.message : "Failed to load module";
      setError(errorMessage);
      toast.error("Failed to load module");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "400px" }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !module) {
    return (
      <Box>
        <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 3 }}>
          <Typography variant="h4" component="h1">
            Module Details
          </Typography>
          <Button variant="outlined" onClick={() => navigate("/modules")} startIcon={<ArrowBackIcon />}>
            Back to List
          </Button>
        </Box>
        <Alert severity="error">{error || "Module not found"}</Alert>
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 4 }}>
        <Box>
          <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
            {module.name}
          </Typography>
          <Box sx={{ display: "flex", alignItems: "center", gap: 1, flexWrap: "wrap" }}>
            <Chip label={module.version} color="primary" size="small" />
            <Chip
              label={module.runtime === "http" ? "HTTP" : "Container"}
              color={module.runtime === "http" ? "primary" : "secondary"}
              size="small"
              icon={module.runtime === "http" ? <HttpIcon /> : <ContainerIcon />}
            />
            <Chip
              label={module.project_id || "global"}
              color={module.project_id ? "default" : "info"}
              size="small"
              variant="outlined"
            />
          </Box>
        </Box>
        <Button variant="outlined" onClick={() => navigate("/modules")} startIcon={<ArrowBackIcon />}>
          Back to List
        </Button>
      </Box>

      <Grid container spacing={3}>
        {/* Basic Information */}
        <Grid item xs={12}>
          <Paper sx={{ p: 3 }} elevation={2}>
            <Typography variant="h6" gutterBottom sx={{ fontWeight: 600, mb: 2 }}>
              Basic Information
            </Typography>
            <Grid container spacing={3}>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Module ID
                </Typography>
                <Typography variant="body2" sx={{ fontFamily: "monospace", wordBreak: "break-all" }}>
                  {module.id}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Runtime
                </Typography>
                <Chip
                  icon={
                    module.runtime === "http" ? (
                      <HttpIcon fontSize="small" />
                    ) : (
                      <ContainerIcon fontSize="small" />
                    )
                  }
                  label={module.runtime === "http" ? "HTTP" : "Container"}
                  color={module.runtime === "http" ? "primary" : "secondary"}
                  variant="outlined"
                />
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Version
                </Typography>
                <Chip label={module.version} color="primary" variant="outlined" />
              </Grid>
              <Grid item xs={12} sm={6} md={3}>
                <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                  Project
                </Typography>
                <Chip
                  label={module.project_id || "global"}
                  color={module.project_id ? "default" : "info"}
                  size="small"
                />
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        {/* Runtime Configuration */}
        {module.runtime === "http" && module.http && (
          <Grid item xs={12}>
            <Accordion defaultExpanded elevation={2}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                  <HttpIcon color="primary" />
                  <Typography variant="h6" sx={{ fontWeight: 500 }}>
                    HTTP Configuration
                  </Typography>
                </Box>
              </AccordionSummary>
              <AccordionDetails>
                <Grid container spacing={3}>
                  <Grid item xs={12} md={4}>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      Method
                    </Typography>
                    <Chip label={module.http.method} color="primary" />
                  </Grid>
                  <Grid item xs={12} md={8}>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      URL
                    </Typography>
                    <Typography variant="body2" sx={{ fontFamily: "monospace", wordBreak: "break-all" }}>
                      {module.http.url}
                    </Typography>
                  </Grid>
                  {module.http.headers && Object.keys(module.http.headers).length > 0 && (
                    <Grid item xs={12} md={6}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Headers
                      </Typography>
                      <AceEditor
                        mode="json"
                        theme="github"
                        value={JSON.stringify(module.http.headers, null, 2)}
                        readOnly
                        width="100%"
                        height="150px"
                        fontSize={13}
                        setOptions={{ showLineNumbers: true, tabSize: 2 }}
                      />
                    </Grid>
                  )}
                  {module.http.query_params && Object.keys(module.http.query_params).length > 0 && (
                    <Grid item xs={12} md={6}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Query Parameters
                      </Typography>
                      <AceEditor
                        mode="json"
                        theme="github"
                        value={JSON.stringify(module.http.query_params, null, 2)}
                        readOnly
                        width="100%"
                        height="150px"
                        fontSize={13}
                        setOptions={{ showLineNumbers: true, tabSize: 2 }}
                      />
                    </Grid>
                  )}
                  {module.http.body_template && Object.keys(module.http.body_template).length > 0 && (
                    <Grid item xs={12}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Body Template
                      </Typography>
                      <AceEditor
                        mode="json"
                        theme="github"
                        value={JSON.stringify(module.http.body_template, null, 2)}
                        readOnly
                        width="100%"
                        height="200px"
                        fontSize={13}
                        setOptions={{ showLineNumbers: true, tabSize: 2 }}
                      />
                    </Grid>
                  )}
                  <Grid item xs={12} md={6}>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      Timeout
                    </Typography>
                    <Typography variant="body1">{module.http.timeout_ms || 30000} ms</Typography>
                  </Grid>
                  <Grid item xs={12} md={6}>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      Retry Count
                    </Typography>
                    <Typography variant="body1">{module.http.retry_count || 3}</Typography>
                  </Grid>
                </Grid>
              </AccordionDetails>
            </Accordion>
          </Grid>
        )}

        {module.runtime === "docker" && module.container_registry && (
          <Grid item xs={12}>
            <Accordion defaultExpanded elevation={2}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                  <ContainerIcon color="secondary" />
                  <Typography variant="h6" sx={{ fontWeight: 500 }}>
                    Container Configuration
                  </Typography>
                </Box>
              </AccordionSummary>
              <AccordionDetails>
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                      Image
                    </Typography>
                    <Typography variant="body2" sx={{ fontFamily: "monospace", wordBreak: "break-all" }}>
                      {module.container_registry.image}
                    </Typography>
                  </Grid>
                  {module.container_registry.command && module.container_registry.command.length > 0 && (
                    <Grid item xs={12}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Command
                      </Typography>
                      <Box sx={{ bgcolor: "background.default", p: 2, borderRadius: 1 }}>
                        {module.container_registry.command.map((cmd, idx) => (
                          <Typography key={idx} variant="body2" sx={{ fontFamily: "monospace" }}>
                            {cmd}
                          </Typography>
                        ))}
                      </Box>
                    </Grid>
                  )}
                  {module.container_registry.env && Object.keys(module.container_registry.env).length > 0 && (
                    <Grid item xs={12}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Environment Variables
                      </Typography>
                      <AceEditor
                        mode="json"
                        theme="github"
                        value={JSON.stringify(module.container_registry.env, null, 2)}
                        readOnly
                        width="100%"
                        height="200px"
                        fontSize={13}
                        setOptions={{ showLineNumbers: true, tabSize: 2 }}
                      />
                    </Grid>
                  )}
                  {(module.container_registry.cpu || module.container_registry.memory) && (
                    <>
                      {module.container_registry.cpu && (
                        <Grid item xs={12} md={6}>
                          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                            CPU
                          </Typography>
                          <Typography variant="body1">{module.container_registry.cpu}</Typography>
                        </Grid>
                      )}
                      {module.container_registry.memory && (
                        <Grid item xs={12} md={6}>
                          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                            Memory
                          </Typography>
                          <Typography variant="body1">{module.container_registry.memory}</Typography>
                        </Grid>
                      )}
                    </>
                  )}
                </Grid>
              </AccordionDetails>
            </Accordion>
          </Grid>
        )}

        {/* Schema Definition */}
        {(module.inputs || module.outputs) && (
          <Grid item xs={12}>
            <Accordion defaultExpanded={false} elevation={2}>
              <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                <Typography variant="h6" sx={{ fontWeight: 500 }}>
                  Schema Definition
                </Typography>
              </AccordionSummary>
              <AccordionDetails>
                <Grid container spacing={3}>
                  {module.inputs && Object.keys(module.inputs).length > 0 && (
                    <Grid item xs={12} md={6}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Inputs (JSON Schema)
                      </Typography>
                      <AceEditor
                        mode="json"
                        theme="github"
                        value={JSON.stringify(module.inputs, null, 2)}
                        readOnly
                        width="100%"
                        height="300px"
                        fontSize={14}
                        showPrintMargin={true}
                        showGutter={true}
                        highlightActiveLine={false}
                        setOptions={{
                          showLineNumbers: true,
                          tabSize: 2,
                        }}
                      />
                    </Grid>
                  )}
                  {module.outputs && Object.keys(module.outputs).length > 0 && (
                    <Grid item xs={12} md={6}>
                      <Typography variant="subtitle2" color="text.secondary" gutterBottom>
                        Outputs (JSON Schema)
                      </Typography>
                      <AceEditor
                        mode="json"
                        theme="github"
                        value={JSON.stringify(module.outputs, null, 2)}
                        readOnly
                        width="100%"
                        height="300px"
                        fontSize={14}
                        showPrintMargin={true}
                        showGutter={true}
                        highlightActiveLine={false}
                        setOptions={{
                          showLineNumbers: true,
                          tabSize: 2,
                        }}
                      />
                    </Grid>
                  )}
                </Grid>
              </AccordionDetails>
            </Accordion>
          </Grid>
        )}
      </Grid>
    </Box>
  );
};

export default ModuleDetails;
