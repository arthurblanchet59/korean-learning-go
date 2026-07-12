import { useState } from "react";

import { adminUpdateUser, resetDatabase } from "../../shared/api/studyApi.js";
import { API_BASE_URL } from "../../app/config.js";

export function AdminPanel({ isMutating, runMutation, token }) {
	const apiRoot = API_BASE_URL.replace(/\/api$/, "");
  const [userId, setUserId] = useState("");
  const [name, setName] = useState("");

  return <fieldset aria-busy={isMutating} className="mutation-surface" disabled={isMutating}><section className="management-section admin-panel">
    <p className="eyebrow">Administration</p><h2>Maintenance</h2>
    <div className="admin-grid">
      <div><h3>Modifier un utilisateur</h3><p>Renseigne l'identifiant du compte non administrateur.</p><form className="inline-form" onSubmit={(event) => { event.preventDefault(); runMutation(() => adminUpdateUser(userId, { name }, token)); }}><input onChange={(event) => setUserId(event.target.value)} placeholder="Identifiant utilisateur" required value={userId} /><input onChange={(event) => setName(event.target.value)} placeholder="Nouveau nom" required value={name} /><button className="primary-button" type="submit">Mettre a jour</button></form></div>
      <div className="danger-zone"><h3>Reinitialiser la base</h3><p>Supprime les donnees d'apprentissage et les comptes non administrateurs. L'admin est conserve.</p><button className="danger-button" onClick={() => window.confirm("Reinitialiser toute la base ?") && runMutation(() => resetDatabase(token))} type="button">Reinitialiser</button></div>
    </div>
      <div className="button-row"><a className="secondary-button" href={`${apiRoot}/swagger/index.html`} rel="noreferrer" target="_blank">Swagger UI</a><a className="secondary-button" href={`${apiRoot}/openapi.json`} rel="noreferrer" target="_blank">OpenAPI JSON</a></div>
  </section></fieldset>;
}
