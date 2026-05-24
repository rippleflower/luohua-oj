export const env = {
  apiBaseUrl: (import.meta.env.VITE_API_BASE_URL ?? "").trim(),
  webBaseUrl: (import.meta.env.VITE_WEB_BASE_URL ?? "http://localhost:5174").trim(),
};
