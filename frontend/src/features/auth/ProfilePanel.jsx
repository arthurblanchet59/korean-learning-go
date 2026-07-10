import { useEffect, useState } from "react";

export function ProfilePanel({ currentUser, onUpdateProfile }) {
  const [form, setForm] = useState({ name: "", email: "", password: "" });
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    setForm({
      name: currentUser?.name ?? "",
      email: currentUser?.email ?? "",
      password: ""
    });
  }, [currentUser]);

  async function handleSubmit(event) {
    event.preventDefault();
    const payload = {
      name: form.name,
      email: form.email
    };
    if (form.password) {
      payload.password = form.password;
    }

    const ok = await onUpdateProfile(payload);
    setSaved(ok);
    if (ok) {
      setForm((current) => ({ ...current, password: "" }));
    }
  }

  function updateField(field, value) {
    setSaved(false);
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
          <input value={form.name} onChange={(event) => updateField("name", event.target.value)} />
        </label>
        <label>
          Email
          <input type="email" value={form.email} onChange={(event) => updateField("email", event.target.value)} />
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
        {saved && <p className="form-success">Profil mis a jour.</p>}
        <button className="primary-button" type="submit">
          Enregistrer
        </button>
      </form>
    </section>
  );
}
