import { useQuery } from "@tanstack/react-query";

import { getContest, listContests } from "./api";

export function useContests() {
  return useQuery({
    queryKey: ["contests"],
    queryFn: listContests,
  });
}

export function useContest(slug: string) {
  return useQuery({
    queryKey: ["contest", slug],
    queryFn: () => getContest(slug),
  });
}
