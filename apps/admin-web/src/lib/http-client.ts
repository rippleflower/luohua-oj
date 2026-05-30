import { env } from "./env";

type RequestOptions = Omit<RequestInit, "body"> & { body?: unknown };

function readCookie(name: string): string {
  if (typeof document === "undefined") {
    return "";
  }
  const cookie = document.cookie.split("; ").find((entry) => entry.startsWith(`${name}=`));
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

export async function getJSON<T>(path: string): Promise<T> {
  const response = await fetch(toUrl(path), {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  if (!response.ok) {
    throw new Error(await readErrorMessage(response));
  }
  return response.json() as Promise<T>;
}

async function sendJSON<T>(method: string, path: string, options: RequestOptions = {}): Promise<T> {
  const csrfToken = readCookie("oj_csrf");
  const response = await fetch(toUrl(path), {
    method,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(csrfToken === "" ? {} : { "X-CSRF-Token": csrfToken }),
      ...options.headers,
    },
    body: options.body === undefined ? undefined : JSON.stringify(options.body),
  });
  if (!response.ok) {
    throw new Error(await readErrorMessage(response));
  }
  return response.json() as Promise<T>;
}

export const postJSON = <T,>(path: string, options: RequestOptions) => sendJSON<T>("POST", path, options);
export const patchJSON = <T,>(path: string, options: RequestOptions) => sendJSON<T>("PATCH", path, options);
export const putJSON = <T,>(path: string, options: RequestOptions) => sendJSON<T>("PUT", path, options);

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const payload = (await response.json()) as { error?: unknown };
    if (typeof payload.error === "string" && payload.error.trim() !== "") {
      return payload.error;
    }
  } catch {
    // ignore invalid or empty error bodies and fall back to status text
  }
  return `request failed: ${response.status}`;
}
