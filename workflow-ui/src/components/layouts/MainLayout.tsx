import {
  Dashboard as DashboardIcon,
  PlayArrow as ExecutionIcon,
  Menu as MenuIcon,
  Extension as ModuleIcon,
  AccountTree as WorkflowIcon,
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
  const [sidebarExpanded, setSidebarExpanded] = useState(false);
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
      setSidebarExpanded(false);
    }
  };

  const currentDrawerWidth = sidebarExpanded ? drawerWidth : drawerWidthMinified;

  const drawer = (
    <Box>
      <Box
        sx={{
          height: 64,
          backgroundColor: (theme) => theme.palette.primary.main,
          display: { xs: "none", md: "block" },
        }}
      />
      <Divider />
      <List>
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
                justifyContent: sidebarExpanded ? "initial" : "center",
                px: 2.5,
              }}
            >
              <ListItemIcon
                sx={{
                  minWidth: 0,
                  mr: sidebarExpanded ? 3 : "auto",
                  justifyContent: "center",
                }}
              >
                {sidebarExpanded ? (
                  item.icon
                ) : (
                  <Tooltip title={item.text} placement="right">
                    {item.icon}
                  </Tooltip>
                )}
              </ListItemIcon>
              <ListItemText
                primary={item.text}
                sx={{ opacity: sidebarExpanded ? 1 : 0 }}
              />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: "flex", minHeight: "100vh" }}>
      <AppBar
        position="fixed"
        sx={{
          width: { md: "100%" },
          left: { md: 0 },
          zIndex: (theme) => theme.zIndex.drawer + 1,
          transition: theme.transitions.create(["width", "left"], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.enteringScreen,
          }),
          boxShadow: "none",
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <Toolbar sx={{ px: 0, minHeight: "64px !important" }}>
          <Box sx={{ 
            display: "flex", 
            alignItems: "center", 
            ml: { md: `${currentDrawerWidth}px` },
            pl: { xs: 2, md: 2 },
            transition: theme.transitions.create("margin-left", {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.enteringScreen,
            }) 
          }}>
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2, display: { md: "none" } }}
            >
              <MenuIcon />
            </IconButton>
            <Logo size={32} />
            <Typography variant="h6" noWrap component="div" sx={{ ml: 1.5 }}>
              Workflow Engine
            </Typography>
          </Box>
          <Box sx={{ flexGrow: 1 }} />
          <Box sx={{ mr: 2 }}>
            <FormControl size="small" sx={{ minWidth: 200 }}>
              <Select
                value={projectId}
                onChange={(e) => setProjectId(e.target.value)}
                sx={{
                  color: "inherit",
                  "& .MuiOutlinedInput-notchedOutline": {
                    borderColor: "rgba(255, 255, 255, 0.23)",
                  },
                  "&:hover .MuiOutlinedInput-notchedOutline": {
                    borderColor: "rgba(255, 255, 255, 0.5)",
                  },
                  "& .MuiSvgIcon-root": {
                    color: "inherit",
                  },
                }}
              >
                {projects.map((project) => (
                  <MenuItem key={project} value={project}>
                    {project}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
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
          p: 3,
          width: { md: `calc(100% - ${currentDrawerWidth}px)` },
          mt: 8,
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
