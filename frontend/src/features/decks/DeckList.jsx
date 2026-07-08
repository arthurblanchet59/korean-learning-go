export function DeckList({ decks }) {
  return (
    <section className="deck-panel" aria-label="Decks disponibles">
      <div className="panel-heading">
        <p className="eyebrow">Decks</p>
        <strong>{decks.length}</strong>
      </div>

      <div className="deck-list">
        {decks.length === 0 ? (
          <article className="deck-item">
            <strong>Aucun deck charge</strong>
            <span>Les decks seront recuperes depuis le backend.</span>
          </article>
        ) : (
          decks.map((deck) => (
            <article className="deck-item" key={deck.id}>
              <strong>{deck.name}</strong>
              <span>{deck.description || "Deck personnel"}</span>
            </article>
          ))
        )}
      </div>
    </section>
  );
}
