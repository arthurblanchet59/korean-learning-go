const navItems = ["Dashboard", "Revision", "Decks", "Cartes", "Stats"];

export function Sidebar({ apiOnline, currentUser, onLogout }) {
  return (
    <aside className="sidebar" aria-label="Navigation principale">
      <div className="brand">
        <span>KL</span>
        <div>
          <p>Korean Learning</p>
          <strong>Study deck</strong>
        </div>
      </div>

      <nav className="nav-list">
        {navItems.map((item, index) => (
          <a className={index === 0 ? "nav-item active" : "nav-item"} href="#" key={item}>
            {item}
          </a>
        ))}
      </nav>

      <section className="side-panel">
        <span>API</span>
        <strong data-online={String(apiOnline)}>{apiOnline ? "Connectee" : "Mode demo"}</strong>
      </section>

      <section className="side-panel user-panel">
        <span>Compte</span>
        <strong>{currentUser?.name ?? "Utilisateur"}</strong>
        <small>{currentUser?.email}</small>
        <button type="button" onClick={onLogout}>
          Deconnexion
        </button>
      </section>
    </aside>
  );
}
