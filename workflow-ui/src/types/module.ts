export interface ApiKeyAuth {
  header: string;
  value: string;
}

export interface BearerAuth {
  token: string;
}

export interface OAuth2Auth {
  token_url: string;
  client_id: string;
  client_secret?: string;
  scope?: string;
}

export interface HttpAuth {
  api_key?: ApiKeyAuth;
  bearer?: BearerAuth;
  oauth2?: OAuth2Auth;
}

export interface HttpModuleSpec {
  method: string;
  url: string;
  headers?: Record<string, string>;
  query_params?: Record<string, string>;
  body_template?: Record<string, unknown>;
  auth?: HttpAuth;
  output_mapping?: Record<string, unknown>;
  timeout_ms?: number;
  retry_count?: number;
}

export interface ContainerRegistryModuleSpec {
  image: string;
  command?: string[];
  env?: Record<string, string>;
  cpu?: string;
  memory?: string;
}

export interface Module {
  id: string;
  project_id?: string;
  name: string;
  version: string;
  runtime: string; // "http" | "docker"
  http?: HttpModuleSpec;
  container_registry?: ContainerRegistryModuleSpec;
  inputs?: Record<string, unknown>;
  outputs?: Record<string, unknown>;
}

export interface RegisterModuleRequest {
  projectId: string;
  module: {
    name: string;
    version: string;
    runtime: string;
    http?: HttpModuleSpec;
    container_registry?: ContainerRegistryModuleSpec;
    inputs?: Record<string, unknown>;
    outputs?: Record<string, unknown>;
  };
}

export interface RegisterModuleResponse {
  module_id: string;
}

export interface GetModuleRequest {
  projectId: string;
  name: string;
  version: string;
}

export interface GetModuleResponse {
  module: Module;
}

export interface ListModulesRequest {
  projectId: string;
}

export interface ListModulesResponse {
  modules: Module[];
}
