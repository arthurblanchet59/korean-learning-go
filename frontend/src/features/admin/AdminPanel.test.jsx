import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { adminUpdateUser, fetchAdminUsers } from "../../shared/api/studyApi.js";
import { AdminPanel } from "./AdminPanel.jsx";

vi.mock("../../shared/api/studyApi.js", () => ({
  adminUpdateUser: vi.fn(),
  fetchAdminUsers: vi.fn(),
  resetDatabase: vi.fn()
}));

describe("AdminPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

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
});
