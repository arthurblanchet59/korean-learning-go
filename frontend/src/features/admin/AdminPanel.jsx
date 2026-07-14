import { useEffect, useState } from "react";

import { API_BASE_URL } from "../../app/config.js";
import { adminUpdateUser, fetchAdminUsers, resetDatabase } from "../../shared/api/studyApi.js";

const emptyForm = { name: "", email: "", password: "" };

export function AdminPanel({ isMutating, runMutation, token }) {
  const apiRoot = API_BASE_URL.replace(/\/api$/, "");
  const [users, setUsers] = useState([]);
  const [selectedId, setSelectedId] = useState("");
  const [form, setForm] = useState(emptyForm);
  const [loadError, setLoadError] = useState("");
  const [saved, setSaved] = useState(false);

  const selectedUser = users.find((user) => user.id === selectedId);

  useEffect(() => {
    let active = true;
    fetchAdminUsers(token).then((result) => {
      if (!active) return;
      if (!result.ok) {
        setLoadError(result.error || "Impossible de charger les utilisateurs.");
        return;
      }

      const nextUsers = result.data ?? [];
      setUsers(nextUsers);
      setLoadError("");
      if (nextUsers.length > 0) selectUser(nextUsers[0]);
    });
    return () => { active = false; };
  }, [token]);

  function selectUser(user) {
    if (!user) return;
    setSelectedId(user.id);
    setForm({ name: user.name, email: user.email, password: "" });
    setSaved(false);
  }

  function updateField(field, value) {
    setSaved(false);
    setForm((current) => ({ ...current, [field]: value }));
  }

  function cancelChanges() {
    if (selectedUser) selectUser(selectedUser);
  }

  async function submit(event) {
    event.preventDefault();
    const payload = { name: form.name.trim(), email: form.email.trim() };
    if (form.password) payload.password = form.password;

    const result = await runMutation(() => adminUpdateUser(selectedId, payload, token));
    if (!result?.ok) return;

    setUsers((current) => current.map((user) => user.id === selectedId ? result.data : user));
    setForm((current) => ({ ...current, name: result.data.name, email: result.data.email, password: "" }));
    setSaved(true);
  }

  async function reset() {
    if (!window.confirm("Réinitialiser toute la base ? Les comptes non administrateurs seront supprimés.")) return;
    const result = await runMutation(() => resetDatabase(token));
    if (result?.ok) {
      setUsers([]);
      setSelectedId("");
      setForm(emptyForm);
      setSaved(false);
    }
  }

  return (
    <fieldset aria-busy={isMutating} className="mutation-surface" disabled={isMutating}>
      <section className="management-section admin-panel">
        <p className="eyebrow">Administration</p>
        <h2>Gestion des utilisateurs</h2>

        {loadError && <p className="form-error">{loadError}</p>}
        {!loadError && users.length === 0 && <p>Aucun compte utilisateur à gérer.</p>}

        {users.length > 0 && (
          <div className="admin-users-layout">
            <div className="admin-user-list" aria-label="Comptes non administrateurs">
              <h3>Utilisateurs</h3>
              {users.map((user) => (
                <button
                  className={user.id === selectedId ? "selected" : ""}
                  key={user.id}
                  onClick={() => selectUser(user)}
                  type="button"
                >
                  <strong>{user.name}</strong>
                  <span>{user.email}</span>
                </button>
              ))}
            </div>

            <form className="admin-user-form" onSubmit={submit}>
              <div className="admin-editor-heading">
                <div>
                  <p className="eyebrow">Compte sélectionné</p>
                  <h3>{selectedUser?.name}</h3>
                </div>
                <span>Utilisateur</span>
              </div>
              <label>Nom
                <input minLength="2" onChange={(event) => updateField("name", event.target.value)} required value={form.name} />
              </label>
              <label>Email
                <input onChange={(event) => updateField("email", event.target.value)} required type="email" value={form.email} />
              </label>
              <label>Nouveau mot de passe
                <input minLength="8" onChange={(event) => updateField("password", event.target.value)} placeholder="Laisser vide pour ne pas le changer" type="password" value={form.password} />
              </label>
              {saved && <p className="form-success">Utilisateur mis à jour.</p>}
              <div className="button-row admin-form-actions">
                <button className="secondary-button" onClick={cancelChanges} type="button">Annuler les modifications</button>
                <button className="primary-button" type="submit">Enregistrer</button>
              </div>
            </form>
          </div>
        )}

        <div className="danger-zone admin-danger-zone">
          <div>
            <h3>Réinitialiser la base</h3>
            <p>Supprime les données d’apprentissage et tous les comptes non administrateurs. Le compte administrateur est conservé.</p>
          </div>
          <button className="danger-button" onClick={reset} type="button">Réinitialiser</button>
        </div>

        <div className="button-row admin-doc-links">
          <a className="secondary-button" href={`${apiRoot}/swagger/index.html`} rel="noreferrer" target="_blank">Swagger UI</a>
          <a className="secondary-button" href={`${apiRoot}/openapi.json`} rel="noreferrer" target="_blank">OpenAPI JSON</a>
        </div>
      </section>
    </fieldset>
  );
}
