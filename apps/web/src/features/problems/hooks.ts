import { useQuery } from "@tanstack/react-query";

import { getProblem, listProblems } from "./api";

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
  });
}
