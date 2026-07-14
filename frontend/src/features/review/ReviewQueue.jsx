const kindLabels = {
  hangul: "Hangeul",
  phrase: "Phrase",
  vocabulary: "Vocabulaire"
};

export function ReviewQueue({ cards, activeIndex, isLoading, onSelect }) {
  return (
    <aside className="queue-panel" aria-label="Cartes à réviser">
      <div className="panel-heading">
        <p className="eyebrow">File de révision</p>
        <strong>{cards.length} cartes</strong>
      </div>

      <div className="queue-list">
        {cards.length === 0 ? (
          <div className="queue-empty">
            {isLoading
              ? "Chargement des cartes dues..."
              : "Aucune carte due aujourd'hui. Profite-en pour parcourir une leçon ou enrichir ta bibliothèque."}
          </div>
        ) : (
          cards.map((card, index) => (
            <button
              className={index === activeIndex ? "queue-card active" : "queue-card"}
              key={card.id}
              type="button"
              onClick={() => onSelect(index)}
            >
              <strong>Carte {index + 1}</strong>
              <span>{kindLabels[card.kind] ?? "Révision"}{card.tags?.[0] ? ` · ${card.tags[0]}` : ""}</span>
            </button>
          ))
        )}
      </div>
    </aside>
  );
}
