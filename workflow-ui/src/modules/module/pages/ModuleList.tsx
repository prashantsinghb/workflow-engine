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
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Add as AddIcon,
  Visibility as ViewIcon,
  Search as SearchIcon,
  Http as HttpIcon,
  Storage as ContainerIcon,
} from "@mui/icons-material";
import { toast } from "react-toastify";
import { moduleApi } from "@/services/client/moduleApi";
import type { Module } from "@/types/module";

const ModuleList = () => {
  const navigate = useNavigate();
  const [modules, setModules] = useState<Module[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(true);
  const projectId = "default-project"; // TODO: Get from context or route

  useEffect(() => {
    loadModules();
  }, []);

  const loadModules = async () => {
    try {
      setLoading(true);
      const response = await moduleApi.listModules({ projectId });
      setModules(response.modules || []);
    } catch (error: unknown) {
      console.error("Failed to load modules", error);
      toast.error("Failed to load modules");
    } finally {
      setLoading(false);
    }
  };

  const filteredModules = modules.filter(
    (module) =>
      module.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      module.runtime?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (module.http?.url?.toLowerCase().includes(searchTerm.toLowerCase())) ||
      (module.container_registry?.image?.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  const getModuleUrl = (module: Module): string => {
    if (module.runtime === "http" && module.http) {
      return module.http.url;
    } else if (module.runtime === "docker" && module.container_registry) {
      return module.container_registry.image;
    }
    return "N/A";
  };

  if (loading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "400px" }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 4 }}>
        <Box>
          <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
            Modules
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Manage your reusable workflow modules
          </Typography>
        </Box>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate("/modules/create")}
          size="large"
        >
          Create Module
        </Button>
      </Box>

      {modules.length > 0 && (
        <Paper sx={{ p: 2, mb: 3, elevation: 2 }}>
          <TextField
            placeholder="Search modules by name, runtime, or URL..."
            size="medium"
            fullWidth
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
          />
        </Paper>
      )}

      {modules.length === 0 ? (
        <Card elevation={3}>
          <CardContent sx={{ textAlign: "center", py: 8 }}>
            <Box sx={{ mb: 2 }}>
              <ContainerIcon sx={{ fontSize: 64, color: "text.secondary", opacity: 0.5 }} />
            </Box>
            <Typography variant="h5" gutterBottom sx={{ fontWeight: 500 }}>
              No Modules Yet
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 4, maxWidth: 400, mx: "auto" }}>
              Create your first module to get started. Modules are reusable components that can be used in workflows.
            </Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => navigate("/modules/create")}
              size="large"
            >
              Create Your First Module
            </Button>
          </CardContent>
        </Card>
      ) : (
        <TableContainer component={Paper} elevation={2}>
          <Table>
            <TableHead>
              <TableRow sx={{ bgcolor: "background.default" }}>
                <TableCell sx={{ fontWeight: 600 }}>Name</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Version</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Runtime</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>URL/Image</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Project</TableCell>
                <TableCell sx={{ fontWeight: 600 }} align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {filteredModules.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 6 }}>
                    <Typography variant="body2" color="text.secondary">
                      No modules match your search.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                filteredModules.map((module) => (
                  <TableRow key={`${module.name}-${module.version}`} hover>
                    <TableCell>
                      <Typography variant="body1" fontWeight="medium">
                        {module.name}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip label={module.version} size="small" color="primary" variant="outlined" />
                    </TableCell>
                    <TableCell>
                      <Chip
                        icon={
                          module.runtime === "http" ? (
                            <HttpIcon fontSize="small" />
                          ) : (
                            <ContainerIcon fontSize="small" />
                          )
                        }
                        label={module.runtime === "http" ? "HTTP" : "Container"}
                        size="small"
                        color={module.runtime === "http" ? "primary" : "secondary"}
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: "monospace",
                          fontSize: "0.75rem",
                          maxWidth: 300,
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                        title={getModuleUrl(module)}
                      >
                        {getModuleUrl(module)}
                      </Typography>
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={module.project_id || "global"}
                        size="small"
                        color={module.project_id ? "default" : "info"}
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell align="right">
                      <IconButton
                        size="small"
                        onClick={() => navigate(`/modules/${module.name}/versions/${module.version}`)}
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
      )}
    </Box>
  );
};

export default ModuleList;
