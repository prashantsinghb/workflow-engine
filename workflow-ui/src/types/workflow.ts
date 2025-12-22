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

export interface ListWorkflowsRequest {
  projectId: string;
}

export interface WorkflowInfo {
  id: string;
  name: string;
  version: string;
  projectId: string;
}

export interface ListWorkflowsResponse {
  workflows: WorkflowInfo[];
}

export interface GetWorkflowRequest {
  projectId: string;
  workflowId: string;
}

export interface GetWorkflowResponse {
  workflow: WorkflowInfo;
  yaml: string;
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
  SUCCESS = "SUCCESS", // From GetExecution enum
  SUCCEEDED = "SUCCEEDED", // From ListExecutions string
  FAILED = "FAILED",
}

export interface GetExecutionResponse {
  state: ExecutionState;
  outputs?: Record<string, unknown>;
  error?: string;
}

export interface ListExecutionsRequest {
  projectId: string;
  workflowId?: string;
}

export interface ExecutionInfo {
  id: string;
  workflowId: string;
  workflowName?: string;
  projectId: string;
  clientRequestId: string;
  state: string;
  error?: string;
}

export interface ListExecutionsResponse {
  executions: ExecutionInfo[];
}

export interface GetDashboardStatsRequest {
  projectId: string;
}

export interface GetDashboardStatsResponse {
  totalWorkflows: number;
  totalExecutions: number;
  runningExecutions: number;
  successRate: number;
}

export interface ExecutionTimelineEvent {
  timestamp: string;
  type: string;
  nodeId?: string;
  executor?: string;
  message?: string;
  durationMs?: number;
  payload?: Record<string, unknown>;
}

export interface ExecutionTimeline {
  executionId: string;
  projectId: string;
  workflowId: string;
  status: string;
  startedAt?: string;
  completedAt?: string;
  timeline: ExecutionTimelineEvent[];
}

