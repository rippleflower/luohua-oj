import { spawn } from "node:child_process";

const processes = [
  { name: "api", command: "go", args: ["run", "./apps/api/cmd/api"] },
  { name: "worker", command: "go", args: ["run", "./apps/judge-worker/cmd/worker"] },
  { name: "web", command: "pnpm", args: ["--filter", "@oj/web", "dev"] },
  { name: "admin", command: "pnpm", args: ["--filter", "@oj/admin-web", "dev"] },
];

const children = processes.map((processConfig) => {
  const child = spawn(processConfig.command, processConfig.args, {
    cwd: process.cwd(),
    env: process.env,
    stdio: "inherit",
  });

  child.on("exit", (code, signal) => {
    const reason = signal ? `signal ${signal}` : `code ${code ?? 0}`;
    console.log(`[dev:${processConfig.name}] exited with ${reason}`);
    shutdown();
  });

  return child;
});

let stopping = false;

function shutdown() {
  if (stopping) {
    return;
  }
  stopping = true;
  for (const child of children) {
    if (!child.killed) {
      child.kill("SIGTERM");
    }
  }
}

process.on("SIGINT", shutdown);
process.on("SIGTERM", shutdown);
