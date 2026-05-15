import { env } from "./env";

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

function toUrl(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  const baseUrl = env.apiBaseUrl.replace(/\/$/, "");
  const route = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${route}`;
}

export async function getJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(toUrl(path), {
    ...init,
    headers: {
      Accept: "application/json",
      ...init?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export async function postJSON<T>(path: string, options: RequestOptions): Promise<T> {
  const response = await fetch(toUrl(path), {
    method: "POST",
    ...options,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...options.headers,
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}
