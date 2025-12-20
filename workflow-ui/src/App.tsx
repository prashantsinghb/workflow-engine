import { FC } from "react";
import { BrowserRouter } from "react-router-dom";
import { ToastContainer } from "react-toastify";
import { ErrorBoundary } from "react-error-boundary";
import { ThemeCustomization } from "./themes";
import Routes from "./routes";
import "react-toastify/dist/ReactToastify.css";

const ErrorFallback = ({ error }: { error: Error }) => {
  return (
    <div role="alert" style={{ padding: "20px", textAlign: "center" }}>
      <h2>Something went wrong:</h2>
      <pre style={{ color: "red" }}>{error.message}</pre>
      <button onClick={() => window.location.reload()}>Reload</button>
    </div>
  );
};

const App: FC = () => {
  return (
    <ErrorBoundary fallbackRender={ErrorFallback} onReset={() => window.location.reload()}>
      <ThemeCustomization>
        <BrowserRouter>
          <ToastContainer
            newestOnTop
            hideProgressBar
            pauseOnFocusLoss={false}
            autoClose={2000}
            draggable
            theme="colored"
            limit={4}
          />
          <Routes />
        </BrowserRouter>
      </ThemeCustomization>
    </ErrorBoundary>
  );
};

export default App;

