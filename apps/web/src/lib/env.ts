type Env = {
  apiBaseUrl: string;
  adminBaseUrl: string;
  submissionsUsername: string;
  demoMode: boolean;
};

const demoModeValue = String(import.meta.env.VITE_DEMO_MODE ?? "")
  .trim()
  .toLowerCase();

export const env: Env = {
  apiBaseUrl: (import.meta.env.VITE_API_BASE_URL ?? "").trim(),
  adminBaseUrl: (import.meta.env.VITE_ADMIN_BASE_URL ?? "").trim(),
  submissionsUsername: (import.meta.env.VITE_SUBMISSIONS_USERNAME ?? "").trim(),
  demoMode: demoModeValue === "true",
};
