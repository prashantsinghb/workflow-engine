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
  Card,
  CardContent,
  CircularProgress,
  Alert,
  Select,
  MenuItem,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Tooltip,
  Grid,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Add as AddIcon,
  Visibility as ViewIcon,
  Search as SearchIcon,
  PlayArrow as PlayIcon,
  MoreVert as MoreVertIcon,
  ViewList as ListViewIcon,
  ViewModule as GridViewIcon,
  Refresh as RefreshIcon,
} from "@mui/icons-material";
import { Formik, Form, Field } from "formik";
import * as Yup from "yup";
import { workflowApi } from "@/services/client/workflowApi";
import { WorkflowInfo } from "@/types/workflow";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";
import Breadcrumbs from "@/components/atoms/Breadcrumbs";

const WorkflowList = () => {
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [workflows, setWorkflows] = useState<WorkflowInfo[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [executeDialogOpen, setExecuteDialogOpen] = useState(false);
  const [selectedWorkflow, setSelectedWorkflow] = useState<WorkflowInfo | null>(null);
  const [executing, setExecuting] = useState(false);
  const [viewMode, setViewMode] = useState<"list" | "grid">("list");

  const loadWorkflows = async () => {
    try {
      setLoading(true);
      setError(null);
      const result = await workflowApi.listWorkflows({ projectId });
      setWorkflows(result.workflows || []);
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "Failed to load workflows";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWorkflows();
  }, []);

  const filteredWorkflows = workflows.filter(
    (workflow) =>
      workflow.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      workflow.id?.toLowerCase().includes(searchTerm.toLowerCase())
  );

  if (loading) {
    return (
      <Box sx={{ textAlign: "center", py: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Alert severity="error">{error}</Alert>
        <Button variant="contained" onClick={() => window.location.reload()} sx={{ mt: 2 }}>
          Retry
        </Button>
      </Box>
    );
  }

  const handleChangePage = (_event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const paginatedWorkflows = filteredWorkflows.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  );

  const handleExecuteClick = (workflow: WorkflowInfo) => {
    setSelectedWorkflow(workflow);
    setExecuteDialogOpen(true);
  };

  const handleExecuteClose = () => {
    setExecuteDialogOpen(false);
    setSelectedWorkflow(null);
  };

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

  return (
    <Box sx={{ p: 3 }}>
      <Breadcrumbs items={[{ label: "Dashboard", path: "/" }, { label: "Workflows" }]} />
      
      <Box sx={{ mb: 3 }}>
        <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", mb: 1 }}>
          <Box>
            <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
              Workflows
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Manage and execute your workflow definitions.
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => navigate("/workflows/create")}
            sx={{
              backgroundColor: "#2e7d32",
              color: "#ffffff",
              textTransform: "none",
              "&:hover": {
                backgroundColor: "#1b5e20",
              },
            }}
          >
            Create Workflow
          </Button>
        </Box>
      </Box>

      {workflows.length > 0 && (
        <Box sx={{ display: "flex", gap: 2, mb: 2, alignItems: "center" }}>
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
          <IconButton onClick={loadWorkflows}>
            <RefreshIcon />
          </IconButton>
        </Box>
      )}

      {workflows.length === 0 ? (
        <Card>
          <CardContent sx={{ textAlign: "center", py: 6 }}>
            <Typography variant="h6" gutterBottom>
              No Workflows Yet
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
              Create your first workflow to get started with the workflow engine.
            </Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => navigate("/workflows/create")}
            >
              Create Your First Workflow
            </Button>
          </CardContent>
        </Card>
      ) : viewMode === "list" ? (
        <>
          <TableContainer component={Paper} sx={{ boxShadow: "none", border: "1px solid #e0e0e0" }}>
            <Table>
              <TableHead>
                <TableRow sx={{ backgroundColor: "#fafafa" }}>
                  <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Version</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Workflow ID</TableCell>
                  <TableCell sx={{ fontWeight: 600 }}>Project</TableCell>
                  <TableCell sx={{ fontWeight: 600, width: 100 }} align="right">Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {paginatedWorkflows.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} align="center" sx={{ py: 4 }}>
                      <Typography variant="body2" color="text.secondary">
                        No workflows match your search.
                      </Typography>
                    </TableCell>
                  </TableRow>
                ) : (
                  paginatedWorkflows.map((workflow) => (
                    <TableRow key={workflow.id} hover>
                      <TableCell>
                        <Typography variant="body1" fontWeight={500}>
                          {workflow.name}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip label={workflow.version} size="small" variant="outlined" />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontFamily: "monospace", color: "text.secondary" }}>
                          {workflow.id?.substring(0, 8)}...
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip label={workflow.projectId || "default"} size="small" />
                      </TableCell>
                      <TableCell align="right">
                        <Box sx={{ display: "flex", gap: 0.5, justifyContent: "flex-end" }}>
                          <Tooltip title="View Details">
                            <IconButton
                              size="small"
                              onClick={() => navigate(`/workflows/${workflow.id}`)}
                              color="primary"
                            >
                              <ViewIcon />
                            </IconButton>
                          </Tooltip>
                          <Tooltip title="Execute Workflow">
                            <IconButton
                              size="small"
                              onClick={() => handleExecuteClick(workflow)}
                              color="success"
                            >
                              <PlayIcon />
                            </IconButton>
                          </Tooltip>
                        </Box>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </TableContainer>
          <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mt: 2 }}>
            <Typography variant="body2" color="text.secondary">
              Showing {page * rowsPerPage + 1}-{Math.min((page + 1) * rowsPerPage, filteredWorkflows.length)} of {filteredWorkflows.length} results
            </Typography>
            <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
              <Typography variant="body2" color="text.secondary">
                Results per page:
              </Typography>
              <Select
                value={rowsPerPage}
                onChange={(e) => {
                  setRowsPerPage(Number(e.target.value));
                  setPage(0);
                }}
                size="small"
                sx={{ minWidth: 80 }}
              >
                <MenuItem value={10}>10</MenuItem>
                <MenuItem value={25}>25</MenuItem>
                <MenuItem value={50}>50</MenuItem>
              </Select>
              <Typography variant="body2" color="text.secondary">
                Page {page + 1} of {Math.ceil(filteredWorkflows.length / rowsPerPage) || 1}
              </Typography>
            </Box>
          </Box>
        </>
      ) : (
        <>
          <Grid container spacing={3}>
            {paginatedWorkflows.length === 0 ? (
              <Grid item xs={12}>
                <Box sx={{ textAlign: "center", py: 4 }}>
                  <Typography variant="body2" color="text.secondary">
                    No workflows match your search.
                  </Typography>
                </Box>
              </Grid>
            ) : (
              paginatedWorkflows.map((workflow) => (
                <Grid item xs={12} sm={6} md={4} key={workflow.id}>
                  <Card sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
                    <CardContent sx={{ flexGrow: 1 }}>
                      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", mb: 2 }}>
                        <Typography variant="h6" component="div" sx={{ fontWeight: 600, flexGrow: 1 }}>
                          {workflow.name}
                        </Typography>
                        <Chip label={workflow.version} size="small" variant="outlined" />
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="caption" color="text.secondary" sx={{ display: "block", mb: 0.5 }}>
                          Workflow ID
                        </Typography>
                        <Typography variant="body2" sx={{ fontFamily: "monospace", color: "text.secondary" }}>
                          {workflow.id?.substring(0, 16)}...
                        </Typography>
                      </Box>
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="caption" color="text.secondary" sx={{ display: "block", mb: 0.5 }}>
                          Project
                        </Typography>
                        <Chip label={workflow.projectId || "default"} size="small" />
                      </Box>
                      <Box sx={{ display: "flex", gap: 1, mt: "auto" }}>
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={<ViewIcon />}
                          onClick={() => navigate(`/workflows/${workflow.id}`)}
                          fullWidth
                        >
                          View
                        </Button>
                        <Button
                          variant="contained"
                          size="small"
                          startIcon={<PlayIcon />}
                          onClick={() => handleExecuteClick(workflow)}
                          fullWidth
                          sx={{
                            backgroundColor: "#2e7d32",
                            "&:hover": {
                              backgroundColor: "#1b5e20",
                            },
                          }}
                        >
                          Execute
                        </Button>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              ))
            )}
          </Grid>
          <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mt: 2 }}>
            <Typography variant="body2" color="text.secondary">
              Showing {page * rowsPerPage + 1}-{Math.min((page + 1) * rowsPerPage, filteredWorkflows.length)} of {filteredWorkflows.length} results
            </Typography>
            <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
              <Typography variant="body2" color="text.secondary">
                Results per page:
              </Typography>
              <Select
                value={rowsPerPage}
                onChange={(e) => {
                  setRowsPerPage(Number(e.target.value));
                  setPage(0);
                }}
                size="small"
                sx={{ minWidth: 80 }}
              >
                <MenuItem value={10}>10</MenuItem>
                <MenuItem value={25}>25</MenuItem>
                <MenuItem value={50}>50</MenuItem>
              </Select>
              <Typography variant="body2" color="text.secondary">
                Page {page + 1} of {Math.ceil(filteredWorkflows.length / rowsPerPage) || 1}
              </Typography>
            </Box>
          </Box>
        </>
      )}

      {/* Execute Workflow Dialog */}
      <Dialog
        open={executeDialogOpen}
        onClose={handleExecuteClose}
        maxWidth="sm"
        fullWidth
      >
        <Formik
          initialValues={{
            clientRequestId: `req-${Date.now()}`,
            inputs: "{}",
          }}
          validationSchema={executionSchema}
          onSubmit={async (values, { setSubmitting }) => {
            if (!selectedWorkflow) return;
            
            try {
              setExecuting(true);
              let inputs = {};
              if (values.inputs && values.inputs.trim() !== "") {
                inputs = JSON.parse(values.inputs);
              }

              const result = await workflowApi.startWorkflow({
                projectId,
                workflowId: selectedWorkflow.id,
                clientRequestId: values.clientRequestId,
                inputs,
              });

              toast.success("Workflow execution started!");
              handleExecuteClose();
              navigate(`/workflows/executions/${result.executionId}`);
            } catch (error: unknown) {
              const errorMessage = error instanceof Error ? error.message : "Failed to start workflow";
              toast.error(errorMessage);
            } finally {
              setExecuting(false);
              setSubmitting(false);
            }
          }}
        >
          {({ values, errors, touched, handleChange, isSubmitting }) => (
            <Form>
              <DialogTitle>
                Execute Workflow: {selectedWorkflow?.name || selectedWorkflow?.id}
              </DialogTitle>
              <DialogContent>
                <Box sx={{ pt: 2 }}>
                  <TextField
                    fullWidth
                    label="Client Request ID"
                    name="clientRequestId"
                    value={values.clientRequestId}
                    onChange={handleChange}
                    error={touched.clientRequestId && !!errors.clientRequestId}
                    helperText={touched.clientRequestId && errors.clientRequestId}
                    margin="normal"
                    required
                  />
                  <TextField
                    fullWidth
                    label="Inputs (JSON)"
                    name="inputs"
                    value={values.inputs}
                    onChange={handleChange}
                    error={touched.inputs && !!errors.inputs}
                    helperText={touched.inputs && errors.inputs || "Enter JSON object for workflow inputs (optional)"}
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
                </Box>
              </DialogContent>
              <DialogActions>
                <Button onClick={handleExecuteClose} disabled={executing || isSubmitting}>
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="contained"
                  disabled={executing || isSubmitting}
                  sx={{
                    bgcolor: "#2e7d32",
                    "&:hover": {
                      bgcolor: "#1b5e20",
                    },
                  }}
                  startIcon={executing || isSubmitting ? <CircularProgress size={16} /> : <PlayIcon />}
                >
                  {executing || isSubmitting ? "Executing..." : "Execute"}
                </Button>
              </DialogActions>
            </Form>
          )}
        </Formik>
      </Dialog>
    </Box>
  );
};

export default WorkflowList;

