import type { Language, ProblemDetail, ProblemSummary } from "@oj/shared";

export type ProblemRowStatus = "UNSOLVED" | "ATTEMPTED" | "SOLVED";

export type ProblemRow = ProblemSummary & {
  status: ProblemRowStatus;
};

export type ProblemFilterState = {
  search: string;
  difficulty: "ALL" | ProblemSummary["difficulty"];
  status: "ALL" | ProblemRowStatus;
  tag: string;
};

export type EditorPreferences = {
  theme: "light" | "vs-dark";
  fontSize: number;
  tabSize: number;
  wordWrap: "on" | "off";
  minimap: boolean;
  lineNumbers: boolean;
};

export const defaultProblemFilterState: ProblemFilterState = {
  search: "",
  difficulty: "ALL",
  status: "ALL",
  tag: "ALL",
};

export const defaultEditorPreferences: EditorPreferences = {
  theme: "light",
  fontSize: 14,
  tabSize: 2,
  wordWrap: "on",
  minimap: false,
  lineNumbers: true,
};

export const editorPreferencesStorageKey = "luooj.problem-editor.preferences";

export function createProblemRows(problems: ProblemSummary[]): ProblemRow[] {
  return problems.map((problem, index) => ({
    ...problem,
    status: ["SOLVED", "ATTEMPTED", "UNSOLVED"][index % 3] as ProblemRowStatus,
  }));
}

export function filterProblemRows(
  problems: ProblemRow[],
  filters: ProblemFilterState,
): ProblemRow[] {
  const query = filters.search.trim().toLowerCase();

  return problems.filter((problem) => {
    if (
      query !== "" &&
      !problem.title.toLowerCase().includes(query) &&
      !problem.slug.toLowerCase().includes(query)
    ) {
      return false;
    }
    if (filters.difficulty !== "ALL" && problem.difficulty !== filters.difficulty) {
      return false;
    }
    if (filters.status !== "ALL" && problem.status !== filters.status) {
      return false;
    }
    if (filters.tag !== "ALL" && !problem.tags.includes(filters.tag)) {
      return false;
    }
    return true;
  });
}

export function collectProblemTags(problems: ProblemSummary[]): string[] {
  const counts = new Map<string, number>();
  for (const problem of problems) {
    for (const tag of problem.tags) {
      counts.set(tag, (counts.get(tag) ?? 0) + 1);
    }
  }

  return [...counts.entries()]
    .sort((left, right) => {
      if (right[1] !== left[1]) {
        return right[1] - left[1];
      }
      return left[0].localeCompare(right[0]);
    })
    .map(([tag]) => tag);
}

export function countProblemStatuses(problems: ProblemRow[]): Record<ProblemRowStatus, number> {
  return problems.reduce<Record<ProblemRowStatus, number>>(
    (accumulator, problem) => {
      accumulator[problem.status] += 1;
      return accumulator;
    },
    { UNSOLVED: 0, ATTEMPTED: 0, SOLVED: 0 },
  );
}

export function deriveProblemTags(problem: ProblemDetail): string[] {
  const tags = problem.metadataJson.tags;
  if (!Array.isArray(tags)) {
    return [];
  }

  return tags.filter((item): item is string => typeof item === "string" && item.trim() !== "");
}

export function defaultSourceForLanguage(language: Language): string {
  switch (language) {
    case "CPP17":
    case "CPP20":
      return [
        "#include <bits/stdc++.h>",
        "using namespace std;",
        "",
        "int main() {",
        "  ios::sync_with_stdio(false);",
        "  cin.tie(nullptr);",
        "",
        "  return 0;",
        "}",
        "",
      ].join("\n");
    case "JAVA17":
      return [
        "import java.io.*;",
        "import java.util.*;",
        "",
        "public class Main {",
        "  public static void main(String[] args) throws Exception {",
        "  }",
        "}",
        "",
      ].join("\n");
    case "PYTHON311":
      return [
        "import sys",
        "",
        "",
        "def solve() -> None:",
        "    pass",
        "",
        "",
        "if __name__ == '__main__':",
        "    solve()",
        "",
      ].join("\n");
  }
}

export function monacoLanguageFor(language: Language): string {
  switch (language) {
    case "CPP17":
    case "CPP20":
      return "cpp";
    case "JAVA17":
      return "java";
    case "PYTHON311":
      return "python";
  }
}

export function loadEditorPreferences(): EditorPreferences {
  if (typeof window === "undefined") {
    return defaultEditorPreferences;
  }

  const raw = window.localStorage.getItem(editorPreferencesStorageKey);
  if (!raw) {
    return defaultEditorPreferences;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<EditorPreferences>;
    return {
      theme: parsed.theme === "vs-dark" ? "vs-dark" : "light",
      fontSize:
        typeof parsed.fontSize === "number" && parsed.fontSize >= 12 && parsed.fontSize <= 24
          ? parsed.fontSize
          : defaultEditorPreferences.fontSize,
      tabSize:
        typeof parsed.tabSize === "number" && parsed.tabSize >= 2 && parsed.tabSize <= 8
          ? parsed.tabSize
          : defaultEditorPreferences.tabSize,
      wordWrap: parsed.wordWrap === "off" ? "off" : "on",
      minimap: Boolean(parsed.minimap),
      lineNumbers: parsed.lineNumbers !== false,
    };
  } catch {
    return defaultEditorPreferences;
  }
}

export function saveEditorPreferences(preferences: EditorPreferences) {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(
    editorPreferencesStorageKey,
    JSON.stringify(preferences),
  );
}
