import { useQuery } from "@tanstack/react-query";

import { getAuthMe, getMeSettings, getMeSummary, listSessions } from "./api";

export function useAuthUser() {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: getAuthMe,
    retry: false,
  });
}

export function useMeSummary() {
  return useQuery({
    queryKey: ["me", "summary"],
    queryFn: getMeSummary,
    retry: false,
  });
}

export function useMeSettings() {
  return useQuery({
    queryKey: ["me", "settings"],
    queryFn: getMeSettings,
    retry: false,
  });
}

export function useSessions() {
  return useQuery({
    queryKey: ["auth", "sessions"],
    queryFn: listSessions,
    retry: false,
  });
}
