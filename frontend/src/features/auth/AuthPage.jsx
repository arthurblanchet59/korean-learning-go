import { useState } from "react";

export function AuthPage({ onLogin, onRegister, isLoading, error }) {
  const [mode, setMode] = useState("login");
  const [form, setForm] = useState({
    name: "",
    email: "admin@korean.local",
    password: "admin123"
  });

  const isRegister = mode === "register";

  async function handleSubmit(event) {
    event.preventDefault();

    const payload = isRegister
      ? { ...form, name: form.name.trim(), email: form.email.trim() }
      : {
          email: form.email.trim(),
          password: form.password
        };

    if (isRegister) {
      await onRegister(payload);
      return;
    }

    await onLogin(payload);
  }

  function updateField(field, value) {
    setForm((current) => ({ ...current, [field]: value }));
  }

  return (
    <main className="auth-shell">
      <section className="auth-panel" aria-label="Authentification">
        <div>
          <p className="eyebrow">Korean Learning</p>
          <h1>{isRegister ? "Créer un compte" : "Connexion"}</h1>
          <p className="auth-copy">
            Connecte-toi pour charger tes decks, tes cartes et ton historique de revision depuis l'API.
          </p>
        </div>

        <div className="auth-tabs" role="tablist" aria-label="Mode d'authentification">
          <button className={!isRegister ? "active" : ""} type="button" onClick={() => setMode("login")}>
            Login
          </button>
          <button className={isRegister ? "active" : ""} type="button" onClick={() => setMode("register")}>
            Register
          </button>
        </div>

        <form className="auth-form" onSubmit={handleSubmit}>
          {isRegister && (
            <label>
              Nom
              <input
                minLength={2}
                name="name"
                onChange={(event) => updateField("name", event.target.value)}
                required
                value={form.name}
              />
            </label>
          )}

          <label>
            Email
            <input
              name="email"
              onChange={(event) => updateField("email", event.target.value)}
              required
              type="email"
              value={form.email}
            />
          </label>

          <label>
            Mot de passe
            <input
              minLength={8}
              name="password"
              onChange={(event) => updateField("password", event.target.value)}
              required
              type="password"
              value={form.password}
            />
          </label>

          {error && <p className="form-error">{error}</p>}

          <button className="primary-button" disabled={isLoading} type="submit">
            {isLoading ? "Chargement..." : isRegister ? "Créer le compte" : "Se connecter"}
          </button>
        </form>
      </section>
    </main>
  );
}
