import { useEffect, useState } from "react";

export function ProfilePanel({ currentUser, onLogout, onUpdateProfile }) {
  const [form, setForm] = useState({ name: "", email: "", password: "" });
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    setForm({
      name: currentUser?.name ?? "",
      email: currentUser?.email ?? "",
      password: ""
    });
  }, [currentUser]);

  async function handleSubmit(event) {
    event.preventDefault();
    setError("");
    const payload = {
      name: form.name.trim(),
      email: form.email.trim()
    };
    if (form.password) {
      payload.password = form.password;
    }

    const ok = await onUpdateProfile(payload);
    setSaved(ok);
    setError(ok ? "" : "La modification du profil a échoué.");
    if (ok) {
      setForm((current) => ({ ...current, password: "" }));
    }
  }

  function updateField(field, value) {
    setSaved(false);
    setError("");
    setForm((current) => ({ ...current, [field]: value }));
  }

  return (
    <section className="deck-panel profile-panel" aria-label="Profil utilisateur">
      <div className="panel-heading">
        <p className="eyebrow">Profil</p>
        <strong>{currentUser?.isAdmin ? "Admin" : "Utilisateur"}</strong>
      </div>

      <form className="profile-form" onSubmit={handleSubmit}>
        <label>
          Nom
          <input required value={form.name} onChange={(event) => updateField("name", event.target.value)} />
        </label>
        <label>
          Email
          <input required type="email" value={form.email} onChange={(event) => updateField("email", event.target.value)} />
        </label>
        <label>
          Nouveau mot de passe
          <input
            minLength={8}
            type="password"
            value={form.password}
            onChange={(event) => updateField("password", event.target.value)}
          />
        </label>
        {saved && <p className="form-success">Profil mis à jour.</p>}
        {error && <p className="form-error">{error}</p>}
        <button className="primary-button" type="submit">
          Modifier mes informations
        </button>
      </form>

      <div className="profile-session">
        <div>
          <strong>Session</strong>
          <p>Déconnecte ce compte de cet appareil.</p>
        </div>
        <button className="secondary-button danger-button" onClick={onLogout} type="button">
          Se déconnecter
        </button>
      </div>
    </section>
  );
}
