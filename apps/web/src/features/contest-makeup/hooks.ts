import { useQuery } from "@tanstack/react-query";

import { getContestMakeupList } from "./api";

export function useContestMakeupList(slug: string) {
  return useQuery({
    queryKey: ["contest", slug, "makeup-list"],
    queryFn: () => getContestMakeupList(slug),
    enabled: slug.trim() !== "",
  });
}
