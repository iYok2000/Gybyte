import axios, {
  type AxiosInstance,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from "axios";
import { API_CONFIG } from "./config";

// Factory + preconfigured axios instances (Req 15.6).
// createAxiosInstance wires a base URL, JSON content type, a shared timeout,
// and request/response interceptors. Feature clients (e.g. auditService) can
// build their own instance and override the timeout as needed.

export function createAxiosInstance(baseURL: string): AxiosInstance {
  const instance = axios.create({
    baseURL,
    timeout: API_CONFIG.TIMEOUT,
    headers: {
      "Content-Type": "application/json",
    },
  });

  // Request interceptor: hook point for auth headers / correlation ids.
  instance.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => config,
    (error) => Promise.reject(error),
  );

  // Response interceptor: pass successful responses through untouched and let
  // callers handle rejections (auditService maps them to AuditError).
  instance.interceptors.response.use(
    (response: AxiosResponse) => response,
    (error) => Promise.reject(error),
  );

  return instance;
}

// Node/BFF API client (other domains).
export const nodeApi = createAxiosInstance(API_CONFIG.NODE_API);

// Go backend client.
export const goApi = createAxiosInstance(API_CONFIG.GO_API);

// Same-origin default client (relative base URL).
export const defaultApi = createAxiosInstance("");
