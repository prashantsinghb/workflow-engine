import React, { useState, useEffect, useMemo } from "react";
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
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import { useNavigate } from "react-router-dom";
import {
  Add as AddIcon,
  Visibility as ViewIcon,
  Search as SearchIcon,
  Http as HttpIcon,
  Storage as ContainerIcon,
  Code as CodeIcon,
  ExpandMore as ExpandMoreIcon,
} from "@mui/icons-material";
import { toast } from "react-toastify";
import { moduleApi } from "@/services/client/moduleApi";
import type { Module } from "@/types/module";
import { useProject } from "@/contexts/ProjectContext";

const ModuleList = () => {
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [modules, setModules] = useState<Module[]>([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [loading, setLoading] = useState(true);

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

  // Compare versions (handles semantic versioning like v1, v2, v1.0, v1.0.1, etc.)
  const compareVersions = (a: string, b: string): number => {
    // Remove 'v' prefix if present
    const normalize = (v: string) => v.replace(/^v/i, "");
    const aNorm = normalize(a);
    const bNorm = normalize(b);

    // Split by dots and compare numerically
    const aParts = aNorm.split(".").map((p) => parseInt(p, 10) || 0);
    const bParts = bNorm.split(".").map((p) => parseInt(p, 10) || 0);

    const maxLength = Math.max(aParts.length, bParts.length);
    for (let i = 0; i < maxLength; i++) {
      const aPart = aParts[i] || 0;
      const bPart = bParts[i] || 0;
      if (aPart > bPart) return -1; // a is newer
      if (aPart < bPart) return 1; // b is newer
    }
    return 0; // equal
  };

  // Group modules by name and sort versions
  const groupedModules = useMemo(() => {
    const groups = new Map<string, Module[]>();
    filteredModules.forEach((module) => {
      const name = module.name || "";
      if (!groups.has(name)) {
        groups.set(name, []);
      }
      groups.get(name)!.push(module);
    });

    // Sort each group by version (latest first)
    groups.forEach((moduleList) => {
      moduleList.sort((a, b) => compareVersions(a.version || "", b.version || ""));
    });

    return Array.from(groups.entries()).map(([name, moduleList]) => ({
      name,
      latest: moduleList[0],
      others: moduleList.slice(1),
    }));
  }, [filteredModules]);

  const getRuntimeConfig = (module: Module): Record<string, unknown> | undefined => {
    return module.runtime_config || (module as any).runtimeConfig;
  };

  const getModuleUrl = (module: Module): string => {
    if (module.runtime === "http" && module.http) {
      return module.http.url;
    } else if (module.runtime === "docker" && module.container_registry) {
      return module.container_registry.image;
    } else {
      const runtimeConfig = getRuntimeConfig(module);
      if (runtimeConfig?.endpoint) {
        // For grpc/go modules, show endpoint from runtime_config
        return String(runtimeConfig.endpoint);
      }
    }
    return "N/A";
  };

  const getModuleEndpoint = (module: Module): string | null => {
    const runtimeConfig = getRuntimeConfig(module);
    if (runtimeConfig?.endpoint) {
      return String(runtimeConfig.endpoint);
    }
    return null;
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
                <TableCell sx={{ fontWeight: 600 }}>URL/Image/Endpoint</TableCell>
                <TableCell sx={{ fontWeight: 600 }}>Project</TableCell>
                <TableCell sx={{ fontWeight: 600 }} align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {groupedModules.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} align="center" sx={{ py: 6 }}>
                    <Typography variant="body2" color="text.secondary">
                      No modules match your search.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                groupedModules.map((group) => (
                  <React.Fragment key={group.name}>
                    {/* Latest version - always visible */}
                    <TableRow hover>
                      <TableCell>
                        <Typography variant="body1" fontWeight="medium">
                          {group.name}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: "flex", alignItems: "center", gap: 0.5 }}>
                          <Chip
                            label={group.latest.version}
                            size="small"
                            color="primary"
                            variant="outlined"
                          />
                          {group.others.length > 0 && (
                            <Chip
                              label="Latest"
                              size="small"
                              color="success"
                              sx={{ fontSize: "0.65rem", height: "18px" }}
                            />
                          )}
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          icon={
                            group.latest.runtime === "http" ? (
                              <HttpIcon fontSize="small" />
                            ) : group.latest.runtime === "go" ? (
                              <CodeIcon fontSize="small" />
                            ) : (
                              <ContainerIcon fontSize="small" />
                            )
                          }
                          label={
                            group.latest.runtime === "http"
                              ? "HTTP"
                              : group.latest.runtime === "go"
                              ? "Go"
                              : group.latest.runtime === "docker"
                              ? "Container"
                              : group.latest.runtime?.toUpperCase() || "Unknown"
                          }
                          size="small"
                          color={
                            group.latest.runtime === "http"
                              ? "primary"
                              : group.latest.runtime === "go"
                              ? "success"
                              : "secondary"
                          }
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>
                        <Box>
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
                            title={getModuleUrl(group.latest)}
                          >
                            {getModuleUrl(group.latest)}
                          </Typography>
                          {(() => {
                            const runtimeConfig = getRuntimeConfig(group.latest);
                            return runtimeConfig?.endpoint && (
                              <Chip
                                label={
                                  runtimeConfig?.protocol
                                    ? `${String(runtimeConfig.protocol).toUpperCase()}: ${String(runtimeConfig.endpoint)}`
                                    : String(runtimeConfig.endpoint)
                                }
                                size="small"
                                variant="outlined"
                                sx={{ mt: 0.5, fontSize: "0.65rem", height: "20px" }}
                              />
                            );
                          })()}
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={group.latest.project_id || "global"}
                          size="small"
                          color={group.latest.project_id ? "default" : "info"}
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() =>
                            navigate(`/modules/${group.latest.name}/versions/${group.latest.version}`)
                          }
                          title="View Details"
                          color="primary"
                        >
                          <ViewIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                    {/* Other versions in accordion */}
                    {group.others.length > 0 && (
                      <TableRow>
                        <TableCell colSpan={6} sx={{ p: 0, border: "none" }}>
                          <Accordion
                            elevation={0}
                            sx={{
                              "&:before": { display: "none" },
                              bgcolor: "background.default",
                            }}
                          >
                            <AccordionSummary
                              expandIcon={<ExpandMoreIcon />}
                              sx={{ px: 2, py: 0.5, minHeight: "auto" }}
                            >
                              <Typography variant="body2" color="text.secondary">
                                {group.others.length} older version{group.others.length > 1 ? "s" : ""}
                              </Typography>
                            </AccordionSummary>
                            <AccordionDetails sx={{ p: 0 }}>
                              <Table size="small">
                                <TableBody>
                                  {group.others.map((module) => (
                  <TableRow key={`${module.name}-${module.version}`} hover>
                    <TableCell>
                                        <Typography variant="body2" color="text.secondary">
                        {module.name}
                      </Typography>
                    </TableCell>
                    <TableCell>
                                        <Chip
                                          label={module.version}
                                          size="small"
                                          color="default"
                                          variant="outlined"
                                        />
                    </TableCell>
                    <TableCell>
                      <Chip
                        icon={
                          module.runtime === "http" ? (
                            <HttpIcon fontSize="small" />
                          ) : module.runtime === "go" ? (
                            <CodeIcon fontSize="small" />
                          ) : (
                            <ContainerIcon fontSize="small" />
                          )
                        }
                        label={
                          module.runtime === "http"
                            ? "HTTP"
                            : module.runtime === "go"
                            ? "Go"
                            : module.runtime === "docker"
                            ? "Container"
                            : module.runtime?.toUpperCase() || "Unknown"
                        }
                        size="small"
                        color={
                          module.runtime === "http"
                            ? "primary"
                            : module.runtime === "go"
                            ? "success"
                            : "secondary"
                        }
                        variant="outlined"
                      />
                    </TableCell>
                    <TableCell>
                                        <Box>
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
                      {(() => {
                        const runtimeConfig = getRuntimeConfig(module);
                        return runtimeConfig?.endpoint && (
                          <Chip
                            label={
                              runtimeConfig?.protocol
                                ? `${String(runtimeConfig.protocol).toUpperCase()}: ${String(runtimeConfig.endpoint)}`
                                : String(runtimeConfig.endpoint)
                            }
                            size="small"
                            variant="outlined"
                            sx={{ mt: 0.5, fontSize: "0.65rem", height: "20px" }}
                          />
                        );
                      })()}
                                        </Box>
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
                                          onClick={() =>
                                            navigate(`/modules/${module.name}/versions/${module.version}`)
                                          }
                        title="View Details"
                        color="primary"
                      >
                        <ViewIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                                  ))}
                                </TableBody>
                              </Table>
                            </AccordionDetails>
                          </Accordion>
                        </TableCell>
                      </TableRow>
                    )}
                  </React.Fragment>
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
