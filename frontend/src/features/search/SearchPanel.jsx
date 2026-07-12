import { useState } from "react";

import { searchAll } from "../../shared/api/studyApi.js";

export function SearchPanel({ token }) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState({ decks: [], cards: [] });
  const [error, setError] = useState("");

  async function search(event) {
    event.preventDefault();
    const result = await searchAll(query, token);
    if (!result.ok) {
      setError(result.error);
      return;
    }
    setError("");
    setResults(result.data);
  }

  return <section className="management-section search-panel">
    <p className="eyebrow">Tous les champs</p><h2>Recherche</h2>
    <form className="search-form" onSubmit={search}><input onChange={(event) => setQuery(event.target.value)} placeholder="Mot coreen, traduction, tag, deck..." required value={query} /><button className="primary-button" type="submit">Rechercher</button></form>
    {error && <p className="form-error">{error}</p>}
    <div className="search-results"><div><h3>Decks · {results.decks.length}</h3>{results.decks.map((deck) => <article className="data-row" key={deck.id}><div><strong>{deck.name}</strong><span>{deck.description}</span></div></article>)}</div><div><h3>Cartes · {results.cards.length}</h3>{results.cards.map((card) => <article className="data-row" key={card.id}><div><strong className="korean-text">{card.korean}</strong><span>{card.translation} · {(card.tags ?? []).join(", ")}</span></div></article>)}</div></div>
  </section>;
}
