import "@mui/material/styles";

declare module "@mui/material/styles" {
  interface Palette {
    sidebar?: {
      main: string;
      light: string;
      contrastText: string;
    };
  }

  interface PaletteOptions {
    sidebar?: {
      main: string;
      light: string;
      contrastText: string;
    };
  }
}

