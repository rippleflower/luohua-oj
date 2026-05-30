import { env } from "./env";

type RequestOptions = Omit<RequestInit, "body"> & {
  body?: unknown;
};

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
  const response = await fetch(toUrl(path), {
    credentials: "include",
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

async function sendJSON<T>(method: string, path: string, options: RequestOptions = {}): Promise<T> {
  const csrfToken = readCookie("oj_csrf");
  const response = await fetch(toUrl(path), {
    method,
    credentials: "include",
    ...options,
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(csrfToken === "" || method === "GET" ? {} : { "X-CSRF-Token": csrfToken }),
      ...options.headers,
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });

  if (!response.ok) {
    throw new Error(`request failed: ${response.status}`);
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
