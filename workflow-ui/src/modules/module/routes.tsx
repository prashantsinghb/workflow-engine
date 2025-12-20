import { Routes, Route } from "react-router-dom";
import ModuleList from "./pages/ModuleList";
import ModuleCreate from "./pages/ModuleCreate";
import ModuleDetails from "./pages/ModuleDetails";

const ModuleRoutes = () => {
  return (
    <Routes>
      <Route path="" element={<ModuleList />} />
      <Route path="create" element={<ModuleCreate />} />
      <Route path=":name/versions/:version" element={<ModuleDetails />} />
    </Routes>
  );
};

export default ModuleRoutes;

