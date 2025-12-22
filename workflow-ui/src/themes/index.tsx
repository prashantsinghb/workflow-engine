import { ThemeProvider, createTheme, CssBaseline } from "@mui/material";
import { ReactNode } from "react";

const theme = createTheme({
  palette: {
    mode: "light",
    primary: {
      main: "#1976d2",
    },
    secondary: {
      main: "#2e7d32",
    },
    background: {
      default: "#f5f5f5",
      paper: "#ffffff",
    },
    sidebar: {
      main: "#ffffff", // White sidebar to match header
      light: "#f5f5f5",
      contrastText: "#000000",
    },
  },
  typography: {
    fontFamily: "'Roboto', sans-serif",
  },
  components: {
    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundColor: "#ffffff",
          color: "#000000",
        },
      },
    },
    MuiListItemButton: {
      styleOverrides: {
        root: {
          "&.Mui-selected": {
            backgroundColor: "rgba(46, 125, 50, 0.15)",
            color: "#2e7d32",
            "&:hover": {
              backgroundColor: "rgba(46, 125, 50, 0.2)",
            },
            "& .MuiListItemIcon-root": {
              color: "#2e7d32",
            },
          },
          "&:hover": {
            backgroundColor: "rgba(0, 0, 0, 0.04)",
          },
        },
      },
    },
    MuiListItemIcon: {
      styleOverrides: {
        root: {
          color: "#000000",
          minWidth: 40,
        },
      },
    },
    MuiListItemText: {
      styleOverrides: {
        primary: {
          color: "#000000",
        },
      },
    },
  },
});

interface ThemeCustomizationProps {
  children: ReactNode;
}

export const ThemeCustomization = ({ children }: ThemeCustomizationProps) => {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
};


