type LogTransport = (entry: Record<string, unknown>) => void;

const defaultTransport: LogTransport = (entry) => {
  console.info(JSON.stringify(entry));
};

let activeTransport: LogTransport = defaultTransport;

export function setLogTransport(transport: LogTransport) {
  activeTransport = transport;
}

export function resetLogTransport() {
  activeTransport = defaultTransport;
}

export function logEvent(entry: Record<string, unknown>) {
  activeTransport({
    ...entry,
    level: "info",
    timestamp: new Date().toISOString(),
  });
}

export function logError(entry: Record<string, unknown>) {
  activeTransport({
    ...entry,
    level: "error",
    timestamp: new Date().toISOString(),
  });
}

export function logRequest(entry: Record<string, unknown>) {
  activeTransport({
    ...entry,
    level: "info",
    type: "request",
    timestamp: new Date().toISOString(),
  });
}
