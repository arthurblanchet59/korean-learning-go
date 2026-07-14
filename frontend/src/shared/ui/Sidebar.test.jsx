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
    expect(screen.getByText("Connectée")).toHaveAttribute("data-online", "true");
		expect(screen.queryByRole("button", { name: "Administration" })).not.toBeInTheDocument();
  });

	it("shows the administration tab only to an admin", () => {
		render(<Sidebar activeView="admin" apiOnline currentUser={{ name: "Admin", email: "admin@example.test", isAdmin: true }} onLogout={vi.fn()} onNavigate={vi.fn()} />);
		expect(screen.getByRole("button", { name: "Administration" })).toHaveClass("active");
	});
});
