import { useQuery } from "@tanstack/react-query";

import { getProblem, getProblemByRouteCode, listProblems } from "./api";

export function useProblems() {
  return useQuery({
    queryKey: ["problems"],
    queryFn: listProblems,
    initialData: [],
  });
}

export function useProblem(slug: string) {
  return useQuery({
    queryKey: ["problem", slug],
    queryFn: () => getProblem(slug),
    enabled: slug.trim() !== "",
  });
}

export function useProblemByRouteCode(routeCode: string) {
  return useQuery({
    queryKey: ["problem", "route-code", routeCode],
    queryFn: () => getProblemByRouteCode(routeCode),
    enabled: routeCode.trim() !== "",
  });
}
