import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LessonsPanel } from "./LessonsPanel.jsx";

describe("LessonsPanel", () => {
  it("shows a simple completion action without exposing a score", () => {
    render(
      <LessonsPanel
        lessons={[{
          id: "lesson-1",
          level: "A1",
          order: 1,
          title: "Premiers mots",
          description: "Introduction",
          content: "OBJECTIF\nLire une phrase.",
          progress: { completed: false, score: 0 }
        }]}
        runMutation={vi.fn()}
        token="token"
      />
    );

    expect(screen.getByText(/Leçon 1/)).toBeInTheDocument();
    expect(screen.queryByLabelText(/Auto-evaluation/i)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Valider la leçon" })).toBeEnabled();
    expect(screen.getByText("Leçon à faire")).toBeInTheDocument();
  });

  it("shows an explicit empty state", () => {
    render(<LessonsPanel lessons={[]} runMutation={vi.fn()} token="token" />);
    expect(screen.getByText(/Aucune leçon/i)).toBeInTheDocument();
  });
});
