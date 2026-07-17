import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { adminUpdateUser, fetchAdminUsers, fetchRAGStatus, reindexRAG } from "../../shared/api/studyApi.js";
import { AdminPanel } from "./AdminPanel.jsx";

vi.mock("../../shared/api/studyApi.js", () => ({
  adminUpdateUser: vi.fn(),
  fetchAdminUsers: vi.fn(),
  fetchRAGStatus: vi.fn(),
  reindexRAG: vi.fn(),
  resetDatabase: vi.fn()
}));

describe("AdminPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    fetchRAGStatus.mockResolvedValue({ ok: true, data: { enabled: true, ready: true, chunkCount: 42, embeddingModel: "embed-v-4-0" } });
  });
  afterEach(cleanup);

  it("selects, cancels and updates a non-admin user through a guided editor", async () => {
    fetchAdminUsers.mockResolvedValue({
      ok: true,
      data: [
        { id: "user-1", name: "Arthur", email: "arthur@example.test", isAdmin: false },
        { id: "user-2", name: "Minji", email: "minji@example.test", isAdmin: false }
      ]
    });
    adminUpdateUser.mockResolvedValue({
      ok: true,
      data: { id: "user-2", name: "Minji Kim", email: "minji@example.test", isAdmin: false }
    });
    const runMutation = (operation) => operation();
    render(<AdminPanel isMutating={false} runMutation={runMutation} token="admin-token" />);

    await waitFor(() => expect(screen.getByRole("button", { name: /Minji/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole("button", { name: /Minji/ }));
    expect(screen.getByLabelText("Nom")).toHaveValue("Minji");

    fireEvent.change(screen.getByLabelText("Nom"), { target: { value: "Valeur temporaire" } });
    fireEvent.click(screen.getByRole("button", { name: "Annuler les modifications" }));
    expect(screen.getByLabelText("Nom")).toHaveValue("Minji");

    fireEvent.change(screen.getByLabelText("Nom"), { target: { value: "Minji Kim" } });
    fireEvent.click(screen.getByRole("button", { name: "Enregistrer" }));
    await waitFor(() => expect(adminUpdateUser).toHaveBeenCalledWith(
      "user-2",
      { name: "Minji Kim", email: "minji@example.test" },
      "admin-token"
    ));
    expect(screen.getByText("Utilisateur mis à jour.")).toBeInTheDocument();
  });

  it("shows and rebuilds the pedagogical index", async () => {
    fetchAdminUsers.mockResolvedValue({ ok: true, data: [] });
    reindexRAG.mockResolvedValue({ ok: true, data: { enabled: true, ready: true, chunkCount: 45, embeddingModel: "embed-v-4-0" } });
    render(<AdminPanel isMutating={false} runMutation={(operation) => operation()} token="admin-token" />);

    expect(await screen.findByText(/42 passage/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Reconstruire l'index" }));
    expect(await screen.findByText(/45 passage/)).toBeInTheDocument();
  });
});
