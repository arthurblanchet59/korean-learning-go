import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LibraryPanel } from "./LibraryPanel.jsx";

const decks = [
  { id: "starter", name: "Essentiels", description: "Premiers mots" },
  { id: "daily", name: "Vie quotidienne", description: "Tous les jours" }
];

const cards = [
  {
    id: "hello",
    deckId: "starter",
    kind: "phrase",
    korean: "안녕하세요",
    translation: "bonjour",
    romanization: "annyeonghaseyo",
    tags: ["salutation"],
    createdAt: "2026-07-10T10:00:00Z"
  },
  {
    id: "water",
    deckId: "daily",
    kind: "vocabulary",
    korean: "물",
    translation: "eau",
    romanization: "mul",
    tags: ["boisson"],
    createdAt: "2026-07-11T10:00:00Z"
  }
];

describe("LibraryPanel", () => {
  it("filters existing cards and opens dedicated create and edit dialogs", () => {
    render(<LibraryPanel cards={cards} decks={decks} isMutating={false} runMutation={vi.fn()} token="token" />);

    expect(screen.getByRole("status")).toHaveTextContent("2 carte(s) affichée(s) sur 2");
    fireEvent.change(screen.getByRole("searchbox", { name: "Rechercher" }), { target: { value: "bonjour" } });
    expect(screen.getByRole("article", { name: "Carte 안녕하세요" })).toBeInTheDocument();
    expect(screen.queryByRole("article", { name: "Carte 물" })).not.toBeInTheDocument();

    fireEvent.change(screen.getByRole("searchbox", { name: "Rechercher" }), { target: { value: "" } });
    fireEvent.change(screen.getByLabelText("Deck"), { target: { value: "daily" } });
    expect(screen.getByRole("article", { name: "Carte 물" })).toBeInTheDocument();
    expect(screen.queryByRole("article", { name: "Carte 안녕하세요" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Ajouter une carte" }));
    const createDialog = screen.getByRole("dialog", { name: "Ajouter une carte" });
    expect(createDialog).toBeInTheDocument();
    expect(within(createDialog).getByLabelText("Deck")).toHaveValue("daily");
    fireEvent.click(screen.getByRole("button", { name: "Fermer" }));

    const waterCard = screen.getByRole("article", { name: "Carte 물" });
    fireEvent.click(within(waterCard).getByRole("button", { name: "Modifier" }));
    const editDialog = screen.getByRole("dialog", { name: "Modifier la carte" });
    expect(editDialog).toBeInTheDocument();
    expect(within(editDialog).getByLabelText("Coréen")).toHaveValue("물");
    expect(within(editDialog).getByLabelText("Traduction")).toHaveValue("eau");
  });
});
