import { env } from "./env";

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
  timeoutMs?: number;
};

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

function readCookie(name: string): string {
  if (typeof document === "undefined") {
    return "";
  }

  const cookie = document.cookie
    .split("; ")
    .find((entry) => entry.startsWith(`${name}=`));
  return cookie ? decodeURIComponent(cookie.slice(name.length + 1)) : "";
}

function toUrl(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  const baseUrl = env.apiBaseUrl.replace(/\/$/, "");
  const route = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${route}`;
}

export async function getJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetchWithTimeout(toUrl(path), {
    credentials: "include",
    ...init,
    headers: {
      Accept: "application/json",
      ...init?.headers,
    },
  });

  if (!response.ok) {
    throw new ApiError(await readErrorMessage(response), response.status);
  }

  return response.json() as Promise<T>;
}

async function sendJSON<T>(method: string, path: string, options: RequestOptions = {}): Promise<T> {
  const csrfToken = readCookie("oj_csrf");
  const response = await fetchWithTimeout(toUrl(path), {
    method,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(csrfToken === "" || method === "GET" ? {} : { "X-CSRF-Token": csrfToken }),
      ...options.headers,
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  }, options.timeoutMs);

  if (!response.ok) {
    throw new ApiError(await readErrorMessage(response), response.status);
  }

  return response.json() as Promise<T>;
}

export async function postJSON<T>(path: string, options: RequestOptions): Promise<T> {
  return sendJSON<T>("POST", path, options);
}

export async function patchJSON<T>(path: string, options: RequestOptions): Promise<T> {
  return sendJSON<T>("PATCH", path, options);
}

export async function putJSON<T>(path: string, options: RequestOptions): Promise<T> {
  return sendJSON<T>("PUT", path, options);
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: unknown };
    if (typeof payload.error === "string" && payload.error.trim() !== "") {
      return payload.error;
    }
  } catch {
    // ignore invalid json error bodies
  }

  return `request failed: ${response.status}`;
}

async function fetchWithTimeout(input: RequestInfo | URL, init?: RequestInit, timeoutMs = 10_000): Promise<Response> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), timeoutMs);

  try {
    return await fetch(input, {
      ...init,
      signal: controller.signal,
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError") {
      throw new ApiError("request timed out", 408);
    }
    throw error;
  } finally {
    window.clearTimeout(timeout);
  }
}
