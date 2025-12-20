import { useState, useEffect } from "react";
import {
  Box,
  Grid,
  Paper,
  Typography,
  Card,
  CardContent,
  Button,
} from "@mui/material";
import {
  AccountTree as WorkflowIcon,
  PlayArrow as ExecutionIcon,
  CheckCircle as SuccessIcon,
} from "@mui/icons-material";
import { useNavigate } from "react-router-dom";

const Dashboard = () => {
  const navigate = useNavigate();
  const [stats] = useState({
    totalWorkflows: 0,
    totalExecutions: 0,
    runningExecutions: 0,
    successRate: 0,
  });

  useEffect(() => {
    // TODO: Fetch real stats from API
    // For now, using placeholder data
  }, []);

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
      value: `${stats.successRate}%`,
      icon: <SuccessIcon sx={{ fontSize: 40 }} />,
      color: "#2e7d32",
    },
  ];

  return (
    <Box>
      <Typography variant="h4" component="h1" gutterBottom>
        Dashboard
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
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

