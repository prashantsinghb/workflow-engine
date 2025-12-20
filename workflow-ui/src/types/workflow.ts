export interface WorkflowDefinition {
  name: string;
  version: string;
  yaml: string;
}

export interface ValidateWorkflowRequest {
  projectId: string;
  workflow: WorkflowDefinition;
}

export interface ValidateWorkflowResponse {
  valid: boolean;
  errors: string[];
}

export interface RegisterWorkflowRequest {
  projectId: string;
  workflow: WorkflowDefinition;
}

export interface RegisterWorkflowResponse {
  workflowId: string;
}

export interface StartWorkflowRequest {
  projectId: string;
  workflowId: string;
  inputs?: Record<string, unknown>;
  clientRequestId: string;
}

export interface StartWorkflowResponse {
  executionId: string;
  state: string;
}

export enum ExecutionState {
  EXECUTION_STATE_UNSPECIFIED = "EXECUTION_STATE_UNSPECIFIED",
  PENDING = "PENDING",
  RUNNING = "RUNNING",
  SUCCESS = "SUCCESS",
  FAILED = "FAILED",
}

export interface GetExecutionResponse {
  state: ExecutionState;
  outputs?: Record<string, unknown>;
  error?: string;
}

