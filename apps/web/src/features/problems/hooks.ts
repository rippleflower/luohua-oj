import { useQuery } from "@tanstack/react-query";

import { listProblems } from "./api";

export function useProblems() {
  return useQuery({
    queryKey: ["problems"],
    queryFn: listProblems,
    initialData: [],
  });
}
