import {
  Dashboard as DashboardIcon,
  PlayArrow as ExecutionIcon,
  Menu as MenuIcon,
  Extension as ModuleIcon,
  AccountTree as WorkflowIcon,
  Notifications as NotificationsIcon,
  KeyboardArrowDown as ArrowDownIcon,
  Add as AddIcon,
  Refresh as RefreshIcon,
  ViewList as ListViewIcon,
  ViewModule as GridViewIcon,
} from "@mui/icons-material";
import {
  AppBar,
  Box,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Tooltip,
  Typography,
  useMediaQuery,
  useTheme,
  Select,
  MenuItem,
  FormControl,
  Button,
  Menu,
  Badge,
} from "@mui/material";
import { ReactNode, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Logo from "../atoms/Logo";
import { useProject } from "@/contexts/ProjectContext";

interface MainLayoutProps {
  children: ReactNode;
}

const drawerWidth = 240;
const drawerWidthMinified = 64;

const menuItems = [
  { text: "Dashboard", icon: <DashboardIcon />, path: "/" },
  { text: "Workflows", icon: <WorkflowIcon />, path: "/workflows" },
  { text: "Modules", icon: <ModuleIcon />, path: "/modules" },
  { text: "Executions", icon: <ExecutionIcon />, path: "/executions" },
];

const MainLayout = ({ children }: MainLayoutProps) => {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [sidebarExpanded, setSidebarExpanded] = useState(true);
  const [createMenuAnchor, setCreateMenuAnchor] = useState<null | HTMLElement>(null);
  const [regionMenuAnchor, setRegionMenuAnchor] = useState<null | HTMLElement>(null);
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("md"));
  const { projectId, setProjectId, projects } = useProject();

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleSidebarMouseEnter = () => {
    if (!isMobile) {
      setSidebarExpanded(true);
    }
  };

  const handleSidebarMouseLeave = () => {
    if (!isMobile) {
      // Keep sidebar expanded by default
    }
  };

  const currentDrawerWidth = sidebarExpanded ? drawerWidth : drawerWidthMinified;

  const handleCreateMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setCreateMenuAnchor(event.currentTarget);
  };

  const handleCreateMenuClose = () => {
    setCreateMenuAnchor(null);
  };

  const handleRegionMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setRegionMenuAnchor(event.currentTarget);
  };

  const handleRegionMenuClose = () => {
    setRegionMenuAnchor(null);
  };

  const drawer = (
    <Box sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
      <Box
        sx={{
          p: 2,
          display: "flex",
          alignItems: "center",
          borderBottom: "1px solid #e0e0e0",
        }}
      >
        <Logo size={24} />
        <Typography variant="h6" sx={{ ml: 1.5, color: "#000000", fontWeight: 600 }}>
          Workflow Engine
        </Typography>
      </Box>
      <Box sx={{ flexGrow: 1, overflow: "auto" }}>
        <List sx={{ pt: 1 }}>
          {menuItems.map((item) => (
            <ListItem key={item.text} disablePadding>
              <ListItemButton
                selected={location.pathname === item.path || location.pathname.startsWith(item.path + "/")}
                onClick={() => {
                  navigate(item.path);
                  if (isMobile) setMobileOpen(false);
                }}
                sx={{
                  minHeight: 48,
                  justifyContent: "initial",
                  px: 2.5,
                  mx: 1,
                  borderRadius: 1,
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 40,
                    mr: 2,
                    justifyContent: "center",
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                <ListItemText primary={item.text} />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      </Box>
      <Box
        sx={{
          p: 2,
          borderTop: "1px solid #e0e0e0",
        }}
      >
        <Typography
          variant="body2"
          sx={{
            color: "rgba(0, 0, 0, 0.7)",
            display: "flex",
            alignItems: "center",
            gap: 1,
            cursor: "pointer",
            "&:hover": {
              color: "#000000",
            },
          }}
        >
          <NotificationsIcon sx={{ fontSize: 18 }} />
          Support
        </Typography>
      </Box>
    </Box>
  );

  return (
    <Box sx={{ display: "flex", minHeight: "100vh" }}>
      <AppBar
        position="fixed"
        sx={{
          width: { md: `calc(100% - ${currentDrawerWidth}px)` },
          left: { md: `${currentDrawerWidth}px` },
          zIndex: (theme) => theme.zIndex.drawer + 1,
          transition: theme.transitions.create(["width", "left"], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
          }),
          boxShadow: "0 1px 3px rgba(0,0,0,0.1)",
          backgroundColor: "#ffffff",
          color: "#000000",
        }}
      >
        <Toolbar sx={{ px: 3, minHeight: "64px !important", justifyContent: "space-between" }}>
          <Box sx={{ display: "flex", alignItems: "center", gap: 3 }}>
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 1, display: { md: "none" } }}
            >
              <MenuIcon />
            </IconButton>
          </Box>
          <Box sx={{ display: "flex", alignItems: "center", gap: 2 }}>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              endIcon={<ArrowDownIcon />}
              onClick={handleCreateMenuOpen}
              sx={{
                backgroundColor: "#2e7d32",
                color: "#ffffff",
                textTransform: "none",
                "&:hover": {
                  backgroundColor: "#1b5e20",
                },
              }}
            >
              Create
            </Button>
            <Menu
              anchorEl={createMenuAnchor}
              open={Boolean(createMenuAnchor)}
              onClose={handleCreateMenuClose}
            >
              <MenuItem onClick={() => { navigate("/workflows/create"); handleCreateMenuClose(); }}>
                Create Workflow
              </MenuItem>
              <MenuItem onClick={() => { navigate("/modules/create"); handleCreateMenuClose(); }}>
                Create Module
              </MenuItem>
            </Menu>
            <Button
              variant="outlined"
              endIcon={<ArrowDownIcon />}
              onClick={handleRegionMenuOpen}
              sx={{
                textTransform: "none",
                borderColor: "#e0e0e0",
                color: "#000000",
                "&:hover": {
                  borderColor: "#bdbdbd",
                  backgroundColor: "rgba(0,0,0,0.04)",
                },
              }}
            >
              {projectId || "dev"}
            </Button>
            <Menu
              anchorEl={regionMenuAnchor}
              open={Boolean(regionMenuAnchor)}
              onClose={handleRegionMenuClose}
            >
              {projects.map((project) => (
                <MenuItem
                  key={project}
                  onClick={() => {
                    setProjectId(project);
                    handleRegionMenuClose();
                  }}
                >
                  {project}
                </MenuItem>
              ))}
            </Menu>
            <IconButton sx={{ color: "#000000" }}>
              <Badge badgeContent={0} color="error">
                <NotificationsIcon />
              </Badge>
            </IconButton>
          </Box>
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{
          width: { md: currentDrawerWidth },
          flexShrink: { md: 0 },
          position: "relative",
        }}
        onMouseEnter={handleSidebarMouseEnter}
        onMouseLeave={handleSidebarMouseLeave}
      >
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{
            keepMounted: true,
          }}
          sx={{
            display: { xs: "block", md: "none" },
            "& .MuiDrawer-paper": {
              boxSizing: "border-box",
              width: drawerWidth,
              backgroundColor: "#ffffff",
              color: "#000000",
              borderRight: "1px solid #e0e0e0",
            },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: "none", md: "block" },
            "& .MuiDrawer-paper": {
              boxSizing: "border-box",
              width: currentDrawerWidth,
              top: 0,
              left: 0,
              zIndex: (theme) => theme.zIndex.drawer,
              transition: theme.transitions.create("width", {
                easing: theme.transitions.easing.sharp,
                duration: theme.transitions.duration.enteringScreen,
              }),
              overflowX: "hidden",
              borderRight: "none",
              backgroundColor: "#ffffff",
              color: "#000000",
              borderRight: "1px solid #e0e0e0",
            },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          width: { md: `calc(100% - ${currentDrawerWidth}px)` },
          mt: 8,
          backgroundColor: "#f5f5f5",
          minHeight: "calc(100vh - 64px)",
          transition: theme.transitions.create(["width", "margin"], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
          }),
        }}
      >
        {children}
      </Box>
    </Box>
  );
};

export default MainLayout;
