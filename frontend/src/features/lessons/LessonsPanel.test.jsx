import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LessonsPanel } from "./LessonsPanel.jsx";

describe("LessonsPanel", () => {
  it("preserves a real score of zero and displays the lesson label", () => {
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
    expect(screen.getByLabelText(/Auto-evaluation/i)).toHaveValue(0);
  });

  it("shows an explicit empty state", () => {
    render(<LessonsPanel lessons={[]} runMutation={vi.fn()} token="token" />);
    expect(screen.getByText(/Aucune lecon/i)).toBeInTheDocument();
  });
});

