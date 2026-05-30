type Env = {
  apiBaseUrl: string;
  adminBaseUrl: string;
  submissionsUsername: string;
};

export const env: Env = {
  apiBaseUrl: (import.meta.env.VITE_API_BASE_URL ?? "").trim(),
  adminBaseUrl: (import.meta.env.VITE_ADMIN_BASE_URL ?? "").trim(),
  submissionsUsername: (import.meta.env.VITE_SUBMISSIONS_USERNAME ?? "").trim(),
};
