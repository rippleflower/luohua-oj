import { beforeEach, describe, expect, it, vi } from "vitest";

describe("problems api", () => {
  beforeEach(() => {
    vi.resetModules();
  });

  it("falls back to local seed only when demo mode is enabled", async () => {
    const mockedGetJSON = vi.fn().mockRejectedValue(new Error("network down"));

    vi.doMock("../../lib/env", () => ({
      env: {
        apiBaseUrl: "",
        adminBaseUrl: "",
        submissionsUsername: "",
        demoMode: true,
      },
    }));
    vi.doMock("../../lib/http-client", async () => {
      const actual = await vi.importActual<typeof import("../../lib/http-client")>(
        "../../lib/http-client",
      );
      return {
        ...actual,
        getJSON: mockedGetJSON,
      };
    });

    const { getProblem, listProblems } = await import("./api");

    const problems = await listProblems();
    const problem = await getProblem("two-sum");

    expect(problems[0]?.title).toBe("Two Sum");
    expect(problem?.samplesJson[0]).toMatchObject({
      input: expect.stringContaining("2 7 11 15"),
      output: "0 1\n",
    });
  });

  it("surfaces a clear error when demo fallback is disabled", async () => {
    const mockedGetJSON = vi.fn().mockRejectedValue(new Error("network down"));

    vi.doMock("../../lib/env", () => ({
      env: {
        apiBaseUrl: "",
        adminBaseUrl: "",
        submissionsUsername: "",
        demoMode: false,
      },
    }));
    vi.doMock("../../lib/http-client", async () => {
      const actual = await vi.importActual<typeof import("../../lib/http-client")>(
        "../../lib/http-client",
      );
      return {
        ...actual,
        getJSON: mockedGetJSON,
      };
    });

    const { listProblems } = await import("./api");

    await expect(listProblems()).rejects.toMatchObject({
      status: 503,
      message: expect.stringContaining("VITE_DEMO_MODE is false"),
    });
  });
});
