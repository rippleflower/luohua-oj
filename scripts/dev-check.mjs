import fs from "node:fs/promises";
import path from "node:path";

const checks = [
  { name: "api", url: "http://127.0.0.1:8080/health", expect: "ok" },
  { name: "web", url: "http://127.0.0.1:5173", expect: "<!doctype html>" },
  { name: "admin", url: "http://127.0.0.1:5174", expect: "<!doctype html>" },
];

for (const check of checks) {
  const response = await fetch(check.url);
  if (!response.ok) {
    throw new Error(`${check.name} check failed: ${response.status}`);
  }
  const body = await response.text();
  if (!body.toLowerCase().includes(check.expect.toLowerCase())) {
    throw new Error(`${check.name} check failed: unexpected response body`);
  }
  console.log(`[ok] ${check.name} -> ${check.url}`);
}

const workerLogPath = path.join(process.cwd(), "tmp", "logs", "judge-worker.log");
const workerLog = await fs.readFile(workerLogPath, "utf8");
if (!workerLog.includes("\"event\":\"worker.started\"") && !workerLog.includes("worker.started")) {
  throw new Error("worker check failed: worker.started not found in tmp/logs/judge-worker.log");
}
console.log("[ok] worker -> startup log detected");
