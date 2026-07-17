import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { previewJournalCorrection } from "../../shared/api/studyApi.js";
import { JournalPanel } from "./JournalPanel.jsx";

vi.mock("../../shared/api/studyApi.js", () => ({
  createJournalEntry: vi.fn(),
  deleteJournalEntry: vi.fn(),
  previewJournalCorrection: vi.fn(),
  updateJournalEntry: vi.fn()
}));

describe("JournalPanel", () => {
  beforeEach(() => vi.clearAllMocks());
  afterEach(cleanup);

  it("shows the backend error when automatic correction is unavailable", async () => {
    previewJournalCorrection.mockResolvedValue({ ok: false, error: "Quota Foundry dépassé." });
    render(<JournalPanel entries={[]} isMutating={false} runMutation={vi.fn()} token="token" />);

    fireEvent.change(screen.getByPlaceholderText("오늘은 무엇을 했어요?"), { target: { value: "저는 학생이에요" } });
    fireEvent.click(screen.getByRole("button", { name: "Vérifier le texte" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Quota Foundry dépassé.");
  });

  it("shows the lessons used by the RAG correction", async () => {
    previewJournalCorrection.mockResolvedValue({
      ok: true,
      data: {
        correctedText: "저는 학생이에요.",
        corrections: [],
        sources: [{ id: "grammar-topic-01", title: "Les particules", level: "A1", excerpt: "은/는 indique le thème." }]
      }
    });
    render(<JournalPanel entries={[]} isMutating={false} runMutation={vi.fn()} token="token" />);

    fireEvent.change(screen.getByPlaceholderText("오늘은 무엇을 했어요?"), { target: { value: "저는 학생이에요" } });
    fireEvent.click(screen.getByRole("button", { name: "Vérifier le texte" }));

    expect(await screen.findByText("Leçons utilisées")).toBeInTheDocument();
    expect(screen.getByText("A1 · Les particules")).toBeInTheDocument();
  });
});
