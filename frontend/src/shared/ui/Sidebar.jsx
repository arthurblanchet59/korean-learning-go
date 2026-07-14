const navItems = [
  ["study", "Révision"],
  ["library", "Bibliothèque"],
  ["lessons", "Leçons"],
  ["journal", "Journal"],
  ["insights", "Progression"],
  ["search", "Recherche"],
  ["profile", "Profil"]
];

export function Sidebar({ activeView, apiOnline, currentUser, onLogout, onNavigate }) {
  return (
    <aside className="sidebar" aria-label="Navigation principale">
      <div className="brand">
        <span>한</span>
        <div>
          <p>Korean Learning</p>
          <strong>Daily study</strong>
        </div>
      </div>

      <nav className="nav-list">
        {[...navItems, ...(currentUser?.isAdmin ? [["admin", "Administration"]] : [])].map(([id, label]) => (
          <button
            className={activeView === id ? "nav-item active" : "nav-item"}
            key={id}
            onClick={() => onNavigate(id)}
            type="button"
          >
            {label}
          </button>
        ))}
      </nav>

      <section className="side-panel">
        <span>API</span>
        <strong data-online={String(apiOnline)}>{apiOnline ? "Connectée" : "Indisponible"}</strong>
      </section>

      <section className="side-panel user-panel">
        <span>{currentUser?.isAdmin ? "Administrateur" : "Compte"}</span>
        <strong>{currentUser?.name ?? "Utilisateur"}</strong>
        <small>{currentUser?.email}</small>
        <button type="button" onClick={onLogout}>Se déconnecter</button>
      </section>
    </aside>
  );
}
