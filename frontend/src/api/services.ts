export type Lifecycle = "development" | "production" | "deprecated";
export type ServiceStatus = "active" | "archived" | "all";

export interface Service {
  id: string;
  serviceKey: string;
  displayName: string;
  description: string;
  ownerTeam: string;
  repositoryUrl: string;
  lifecycle: Lifecycle;
  createdAt: string;
  updatedAt: string;
  archivedAt: string | null;
}

export interface ServiceWriteInput {
  displayName: string;
  description: string;
  ownerTeam: string;
  repositoryUrl: string;
  lifecycle: Lifecycle;
}

export interface ServiceCreateInput extends ServiceWriteInput {
  serviceKey: string;
}

export interface ServicePage {
  items: Service[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface ApiErrorDetail {
  field: string;
  message: string;
}

interface ApiErrorEnvelope {
  error: {
    code: string;
    message: string;
    requestId: string;
    details: ApiErrorDetail[];
  };
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly requestId: string;
  readonly details: ApiErrorDetail[];

  constructor(status: number, envelope: ApiErrorEnvelope) {
    super(envelope.error.message);
    this.name = "ApiError";
    this.status = status;
    this.code = envelope.error.code;
    this.requestId = envelope.error.requestId;
    this.details = envelope.error.details;
  }
}

export async function listServices(
  status: ServiceStatus,
  limit: number,
  offset: number,
  signal?: AbortSignal,
): Promise<ServicePage> {
  const query = new URLSearchParams({
    status,
    limit: String(limit),
    offset: String(offset),
  });
  return request<ServicePage>(`/api/v1/services?${query.toString()}`, {
    signal,
  });
}

export function getService(id: string, signal?: AbortSignal): Promise<Service> {
  return request<Service>(`/api/v1/services/${encodeURIComponent(id)}`, {
    signal,
  });
}

export function createService(input: ServiceCreateInput): Promise<Service> {
  return request<Service>("/api/v1/services", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function updateService(
  id: string,
  input: ServiceWriteInput,
): Promise<Service> {
  return request<Service>(`/api/v1/services/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export async function archiveService(id: string): Promise<void> {
  await request<void>(`/api/v1/services/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}

async function request<T>(url: string, init: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      Accept: "application/json",
      ...init.headers,
    },
  });

  if (!response.ok) {
    const envelope = await parseErrorEnvelope(response);
    throw new ApiError(response.status, envelope);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

async function parseErrorEnvelope(
  response: Response,
): Promise<ApiErrorEnvelope> {
  try {
    const value: unknown = await response.json();
    if (isErrorEnvelope(value)) {
      return value;
    }
  } catch {
    // The fallback below keeps transport failures safe and predictable.
  }
  return {
    error: {
      code: "unexpected_response",
      message: "The server returned an unexpected response.",
      requestId: response.headers.get("X-Request-ID") ?? "",
      details: [],
    },
  };
}

function isErrorEnvelope(value: unknown): value is ApiErrorEnvelope {
  if (!isRecord(value) || !isRecord(value.error)) {
    return false;
  }
  const error = value.error;
  return (
    typeof error.code === "string" &&
    typeof error.message === "string" &&
    typeof error.requestId === "string" &&
    Array.isArray(error.details) &&
    error.details.every(
      (detail) =>
        isRecord(detail) &&
        typeof detail.field === "string" &&
        typeof detail.message === "string",
    )
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
