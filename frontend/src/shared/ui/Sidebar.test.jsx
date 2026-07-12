import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Sidebar } from "./Sidebar.jsx";

describe("Sidebar", () => {
  it("uses the correct Lessons label and keeps API status independent", () => {
    render(
      <Sidebar
        activeView="lessons"
        apiOnline
        currentUser={{ name: "Arthur", email: "arthur@example.test", isAdmin: false }}
        onLogout={vi.fn()}
        onNavigate={vi.fn()}
      />
    );

    expect(screen.getByRole("button", { name: "Leçons" })).toHaveClass("active");
    expect(screen.getByText("Connectee")).toHaveAttribute("data-online", "true");
  });
});

