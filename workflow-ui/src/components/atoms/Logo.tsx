import { Box } from "@mui/material";

const Logo = ({ size = 40 }: { size?: number }) => {
  return (
    <Box
      component="svg"
      width={size}
      height={size}
      viewBox="0 0 64 64"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      sx={{ display: "inline-block", verticalAlign: "middle" }}
    >
      <rect x="8" y="8" width="48" height="48" rx="8" fill="#2e7d32" />
      <path d="M20 32H44" stroke="white" strokeWidth="3" strokeLinecap="round" />
      <path d="M32 20V44" stroke="white" strokeWidth="3" strokeLinecap="round" />
      <circle cx="32" cy="32" r="5" fill="white" />
    </Box>
  );
};

export default Logo;

