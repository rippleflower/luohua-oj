import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { SubmissionForm } from "./submission-form";

describe("SubmissionForm", () => {
  it("submits and emits field changes", () => {
    const onFieldChange = vi.fn();
    const onSubmit = vi.fn((event) => event.preventDefault());

    render(
      <LocaleProvider>
        <SubmissionForm
          form={{
            userId: "",
            problemId: "",
            language: "CPP17",
            source: "int main(){return 0;}",
          }}
          isSubmitting={false}
          onFieldChange={onFieldChange}
          onSubmit={onSubmit}
        />
      </LocaleProvider>,
    );

    fireEvent.change(screen.getByLabelText("用户 ID"), { target: { value: "user-1" } });
    fireEvent.click(screen.getByRole("button", { name: "创建提交" }));

    expect(onFieldChange).toHaveBeenCalledWith("userId", "user-1");
    expect(onSubmit).toHaveBeenCalled();
  });
});
