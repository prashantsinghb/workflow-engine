import { useState, useEffect } from "react";
import {
  Box,
  Grid,
  Paper,
  Typography,
  Card,
  CardContent,
  Button,
  CircularProgress,
} from "@mui/material";
import {
  AccountTree as WorkflowIcon,
  PlayArrow as ExecutionIcon,
  CheckCircle as SuccessIcon,
} from "@mui/icons-material";
import { useNavigate } from "react-router-dom";
import { workflowApi } from "@/services/client/workflowApi";
import { toast } from "react-toastify";
import { useProject } from "@/contexts/ProjectContext";
import Breadcrumbs from "@/components/atoms/Breadcrumbs";

const Dashboard = () => {
  const navigate = useNavigate();
  const { projectId } = useProject();
  const [stats, setStats] = useState({
    totalWorkflows: 0,
    totalExecutions: 0,
    runningExecutions: 0,
    successRate: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        setLoading(true);
        const response = await workflowApi.getDashboardStats({ projectId });
        setStats({
          totalWorkflows: response.totalWorkflows || 0,
          totalExecutions: response.totalExecutions || 0,
          runningExecutions: response.runningExecutions || 0,
          successRate: response.successRate || 0,
        });
      } catch (error: unknown) {
        const errorMessage = error instanceof Error ? error.message : "Failed to fetch dashboard stats";
        toast.error(errorMessage);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
    // Refresh stats every 30 seconds
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, [projectId]);

  const statCards = [
    {
      title: "Total Workflows",
      value: stats.totalWorkflows,
      icon: <WorkflowIcon sx={{ fontSize: 40 }} />,
      color: "#1976d2",
      action: () => navigate("/workflows"),
    },
    {
      title: "Total Executions",
      value: stats.totalExecutions,
      icon: <ExecutionIcon sx={{ fontSize: 40 }} />,
      color: "#ed6c02",
      action: () => navigate("/executions"),
    },
    {
      title: "Running",
      value: stats.runningExecutions,
      icon: <ExecutionIcon sx={{ fontSize: 40 }} />,
      color: "#2e7d32",
      action: () => navigate("/executions?status=RUNNING"),
    },
    {
      title: "Success Rate",
      value: `${stats.successRate.toFixed(1)}%`,
      icon: <SuccessIcon sx={{ fontSize: 40 }} />,
      color: "#2e7d32",
    },
  ];

  if (loading) {
    return (
      <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "400px" }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ p: 3 }}>
      <Breadcrumbs items={[{ label: "Dashboard" }]} />
      <Typography variant="h4" component="h1" sx={{ fontWeight: 600, mb: 0.5 }}>
        Dashboard
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
        Overview of your workflow engine
      </Typography>

      <Grid container spacing={3}>
        {statCards.map((card, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <Card
              sx={{
                height: "100%",
                cursor: card.action ? "pointer" : "default",
                transition: "transform 0.2s, box-shadow 0.2s",
                "&:hover": card.action
                  ? {
                      transform: "translateY(-4px)",
                      boxShadow: 4,
                    }
                  : {},
              }}
              onClick={card.action}
            >
              <CardContent>
                <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
                  <Box sx={{ color: card.color, mr: 2 }}>{card.icon}</Box>
                  <Box sx={{ flexGrow: 1 }}>
                    <Typography variant="h4" component="div">
                      {card.value}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {card.title}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}

        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3, height: "100%" }}>
            <Typography variant="h6" gutterBottom>
              Quick Actions
            </Typography>
            <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 2 }}>
              <Button
                variant="contained"
                fullWidth
                onClick={() => navigate("/workflows/create")}
                sx={{ justifyContent: "flex-start", py: 1.5 }}
              >
                Create New Workflow
              </Button>
              <Button
                variant="outlined"
                fullWidth
                onClick={() => navigate("/workflows")}
                sx={{ justifyContent: "flex-start", py: 1.5 }}
              >
                View All Workflows
              </Button>
              <Button
                variant="outlined"
                fullWidth
                onClick={() => navigate("/executions")}
                sx={{ justifyContent: "flex-start", py: 1.5 }}
              >
                View All Executions
              </Button>
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3, height: "100%" }}>
            <Typography variant="h6" gutterBottom>
              Recent Activity
            </Typography>
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 4 }}>
                No recent activity. Start by creating your first workflow!
              </Typography>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;

