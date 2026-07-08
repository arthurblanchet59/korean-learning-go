export function ReviewQueue({ cards, activeIndex, onSelect }) {
  return (
    <aside className="queue-panel" aria-label="Cartes a reviser">
      <div className="panel-heading">
        <p className="eyebrow">File de revision</p>
        <strong>{cards.length} cartes</strong>
      </div>

      <div className="queue-list">
        {cards.length === 0 ? (
          <div className="queue-empty">
            Lance le backend pour charger les cartes dues depuis l'API.
          </div>
        ) : (
          cards.map((card, index) => (
            <button
              className={index === activeIndex ? "queue-card active" : "queue-card"}
              key={card.id}
              type="button"
              onClick={() => onSelect(index)}
            >
              <strong>{card.korean}</strong>
              <span>{card.translation}</span>
            </button>
          ))
        )}
      </div>
    </aside>
  );
}
