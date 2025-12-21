import { Routes, Route } from "react-router-dom";
import WorkflowList from "./pages/WorkflowList";
import WorkflowCreate from "./pages/WorkflowCreate";
import WorkflowDetails from "./pages/WorkflowDetails";
import ExecutionDetails from "./pages/ExecutionDetails";

const WorkflowRoutes = () => {
  return (
    <Routes>
      <Route path="" element={<WorkflowList />} />
      <Route path="create" element={<WorkflowCreate />} />
      <Route path=":workflowId" element={<WorkflowDetails />} />
      <Route path="executions/:executionId" element={<ExecutionDetails />} />
    </Routes>
  );
};

export default WorkflowRoutes;


