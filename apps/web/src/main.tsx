import { QueryClientProvider } from "@tanstack/react-query";
import React from "react";
import ReactDOM from "react-dom/client";

import { App } from "./app";
import { LocaleProvider } from "./lib/locale";
import { queryClient } from "./lib/query-client";
import "./styles.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <LocaleProvider>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </LocaleProvider>
  </React.StrictMode>,
);
