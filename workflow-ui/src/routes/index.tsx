import { Routes as RouterRoutes, Route } from "react-router-dom";
import MainLayout from "@/components/layouts/MainLayout";
import WorkflowRoutes from "../modules/workflow/routes";
import ModuleRoutes from "../modules/module/routes";
import Dashboard from "../modules/workflow/pages/Dashboard";
import ExecutionList from "../modules/workflow/pages/ExecutionList";

const Routes = () => {
  return (
    <MainLayout>
      <RouterRoutes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/workflows/*" element={<WorkflowRoutes />} />
        <Route path="/modules/*" element={<ModuleRoutes />} />
        <Route path="/executions" element={<ExecutionList />} />
      </RouterRoutes>
    </MainLayout>
  );
};

export default Routes;

