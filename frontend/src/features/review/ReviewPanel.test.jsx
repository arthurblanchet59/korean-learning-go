import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ReviewPanel } from "./ReviewPanel.jsx";

describe("ReviewPanel", () => {
  it("explains each scheduling choice after revealing the answer", () => {
    render(<ReviewPanel activeIndex={0} card={{ id: "card", korean: "집", translation: "maison", romanization: "jip" }} isLoading={false} onAnswer={vi.fn()} onCheck={vi.fn()} totalCards={1} />);
    fireEvent.click(screen.getByRole("button", { name: /Afficher sans répondre/i }));

    expect(screen.getByRole("button", { name: /À revoir/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Avec hésitation/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Bien retenue/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Maîtrisée/i })).toBeInTheDocument();
    expect(screen.getByText("Je ne savais pas")).toBeInTheDocument();
    expect(screen.getByText("Réponse immédiate")).toBeInTheDocument();
  });
});
