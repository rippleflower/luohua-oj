import { beforeEach, describe, expect, it } from "vitest";

import {
  getSubmissionHistoryItem,
  listSubmissionHistory,
  recordSubmissionHistory,
} from "./history";

describe("submission history", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("records and lists submissions", () => {
    recordSubmissionHistory({
      id: "sub-1",
      userId: "user-1",
      problemId: "problem-1",
      language: "CPP17",
      sourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
      status: "PENDING",
    });

    expect(listSubmissionHistory()).toHaveLength(1);
    expect(getSubmissionHistoryItem("sub-1")?.id).toBe("sub-1");
  });
});
