import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ProfilePanel } from "./ProfilePanel.jsx";

describe("ProfilePanel", () => {
  it("updates the current user and offers logout", async () => {
    const onLogout = vi.fn();
    const onUpdateProfile = vi.fn().mockResolvedValue(true);
    render(
      <ProfilePanel
        currentUser={{ name: "Arthur", email: "arthur@example.test", isAdmin: false }}
        onLogout={onLogout}
        onUpdateProfile={onUpdateProfile}
      />
    );

    fireEvent.change(screen.getByLabelText("Nom"), { target: { value: "Arthur B." } });
    fireEvent.change(screen.getByLabelText("Nouveau mot de passe"), { target: { value: "nouveau-secret" } });
    fireEvent.click(screen.getByRole("button", { name: "Modifier mes informations" }));

    await waitFor(() => expect(onUpdateProfile).toHaveBeenCalledWith({
      name: "Arthur B.",
      email: "arthur@example.test",
      password: "nouveau-secret"
    }));
    expect(screen.getByText("Profil mis à jour.")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Se déconnecter" }));
    expect(onLogout).toHaveBeenCalledOnce();
  });
});
