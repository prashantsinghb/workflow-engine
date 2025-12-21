import { apiClient } from "./api";
import type {
  ValidateWorkflowRequest,
  ValidateWorkflowResponse,
  RegisterWorkflowRequest,
  RegisterWorkflowResponse,
  ListWorkflowsRequest,
  ListWorkflowsResponse,
  GetWorkflowRequest,
  GetWorkflowResponse,
  StartWorkflowRequest,
  StartWorkflowResponse,
  GetExecutionResponse,
} from "@/types/workflow";

export const workflowApi = {
  validateWorkflow: async (request: ValidateWorkflowRequest): Promise<ValidateWorkflowResponse> => {
    const response = await apiClient.instance.post<ValidateWorkflowResponse>(
      `/v1/projects/${request.projectId}/workflows:validate`,
      { workflow: request.workflow }
    );
    return response.data;
  },

  registerWorkflow: async (request: RegisterWorkflowRequest): Promise<RegisterWorkflowResponse> => {
    const response = await apiClient.instance.post<RegisterWorkflowResponse>(
      `/v1/projects/${request.projectId}/workflows`,
      { workflow: request.workflow }
    );
    return response.data;
  },

  listWorkflows: async (request: ListWorkflowsRequest): Promise<ListWorkflowsResponse> => {
    const response = await apiClient.instance.get<ListWorkflowsResponse>(
      `/v1/projects/${request.projectId}/workflows`
    );
    return response.data;
  },

  getWorkflow: async (request: GetWorkflowRequest): Promise<GetWorkflowResponse> => {
    const response = await apiClient.instance.get<GetWorkflowResponse>(
      `/v1/projects/${request.projectId}/workflows/${request.workflowId}`
    );
    return response.data;
  },

  startWorkflow: async (request: StartWorkflowRequest): Promise<StartWorkflowResponse> => {
    const response = await apiClient.instance.post<StartWorkflowResponse>(
      `/v1/projects/${request.projectId}/executions`,
      {
        workflowId: request.workflowId,
        inputs: request.inputs || {},
        clientRequestId: request.clientRequestId,
      }
    );
    return response.data;
  },

  getExecution: async (projectId: string, executionId: string): Promise<GetExecutionResponse> => {
    const response = await apiClient.instance.get<GetExecutionResponse>(
      `/v1/projects/${projectId}/executions/${executionId}`
    );
    return response.data;
  },
};


