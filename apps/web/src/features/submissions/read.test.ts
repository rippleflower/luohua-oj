import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../../lib/env", () => ({
  env: {
    apiBaseUrl: "",
    adminBaseUrl: "",
    submissionsUsername: "",
    demoMode: true,
  },
}));

import { recordSubmissionHistory } from "./history";
import { getSubmission, listSubmissions } from "./read";

describe("getSubmission", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("reads from local history in fallback mode", async () => {
    recordSubmissionHistory({
      id: "sub-1",
      userId: "user-1",
      problemId: "problem-1",
      language: "CPP17",
      sourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
      status: "PENDING",
    });

    await expect(getSubmission("sub-1")).resolves.toMatchObject({
      id: "sub-1",
    });
  });

  it("lists local history in fallback mode", async () => {
    recordSubmissionHistory({
      id: "sub-2",
      userId: "user-1",
      problemId: "problem-2",
      language: "CPP17",
      sourceObjectKey: "submissions/2026/05/sub-2/source.cpp.zst",
      status: "PENDING",
    });

    await expect(listSubmissions("demo")).resolves.toMatchObject({
      items: [expect.objectContaining({ id: "sub-2" })],
      total: 1,
      page: 1,
    });
  });

  it("throws a clear error when demo fallback is disabled", async () => {
    vi.resetModules();
    vi.doMock("../../lib/env", () => ({
      env: {
        apiBaseUrl: "",
        adminBaseUrl: "",
        submissionsUsername: "",
        demoMode: false,
      },
    }));

    const { getSubmission: getSubmissionWithRealMode } = await import("./read");

    await expect(getSubmissionWithRealMode("sub-1")).rejects.toMatchObject({
      status: 503,
      message: expect.stringContaining("VITE_DEMO_MODE is false"),
    });
  });
});
