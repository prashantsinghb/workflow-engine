import { apiClient } from "./api";
import type {
  RegisterModuleRequest,
  RegisterModuleResponse,
  GetModuleRequest,
  GetModuleResponse,
  ListModulesRequest,
  ListModulesResponse,
} from "@/types/module";

export const moduleApi = {
  registerModule: async (request: RegisterModuleRequest): Promise<RegisterModuleResponse> => {
    const response = await apiClient.instance.post<RegisterModuleResponse>(
      `/v1/projects/${request.projectId}/modules`,
      {
        name: request.module.name,
        version: request.module.version,
        runtime: request.module.runtime,
        ...(request.module.http ? { http: request.module.http } : {}),
        ...(request.module.container_registry ? { container_registry: request.module.container_registry } : {}),
        inputs: request.module.inputs || {},
        outputs: request.module.outputs || {},
      }
    );
    return response.data;
  },

  getModule: async (request: GetModuleRequest): Promise<GetModuleResponse> => {
    const response = await apiClient.instance.get<GetModuleResponse>(
      `/v1/projects/${request.projectId}/modules/${request.name}/versions/${request.version}`
    );
    return response.data;
  },

  listModules: async (request: ListModulesRequest): Promise<ListModulesResponse> => {
    const response = await apiClient.instance.get<ListModulesResponse>(
      `/v1/projects/${request.projectId}/modules`
    );
    return response.data;
  },
};
