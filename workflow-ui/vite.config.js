import react from "@vitejs/plugin-react-swc";
import path from "path";
import { defineConfig, loadEnv } from "vite";
import checker from "vite-plugin-checker";
import svgr from "vite-plugin-svgr";
import viteTsconfigPaths from "vite-tsconfig-paths";
export default defineConfig(function (_a) {
    var mode = _a.mode;
    var env = loadEnv(mode, process.cwd(), "");
    var port = parseInt(env.VITE_PORT || "3000", 10);
    return {
        plugins: [
            react(),
            viteTsconfigPaths(),
            svgr(),
            checker({
                overlay: { initialIsOpen: false },
                typescript: true,
                eslint: {
                    lintCommand: 'eslint "./src/**/*.{ts,tsx}"',
                },
            }),
        ],
        server: {
            open: true,
            port: port,
            proxy: {
                "/api": {
                    target: env.VITE_API_BASE_URL || "http://localhost:8081",
                    changeOrigin: true,
                    rewrite: function (path) { return path.replace(/^\/api/, ""); },
                },
            },
        },
        preview: { port: port },
        build: { sourcemap: false },
        resolve: {
            alias: {
                "@": path.resolve(__dirname, "./src"),
            },
        },
    };
});
