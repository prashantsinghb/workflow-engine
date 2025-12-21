import { createContext, useContext, useState, ReactNode } from "react";

interface ProjectContextType {
  projectId: string;
  setProjectId: (projectId: string) => void;
  projects: string[];
}

const ProjectContext = createContext<ProjectContextType | undefined>(undefined);

export const ProjectProvider = ({ children }: { children: ReactNode }) => {
  // TODO: Replace with API call to fetch user's projects
  const [projects] = useState<string[]>(["default-project"]);
  const [projectId, setProjectId] = useState<string>("default-project");

  return (
    <ProjectContext.Provider value={{ projectId, setProjectId, projects }}>
      {children}
    </ProjectContext.Provider>
  );
};

export const useProject = () => {
  const context = useContext(ProjectContext);
  if (context === undefined) {
    throw new Error("useProject must be used within a ProjectProvider");
  }
  return context;
};

