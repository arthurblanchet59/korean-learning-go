const ratingOptions = [
  { label: "Encore", value: "again" },
  { label: "Difficile", value: "hard" },
  { label: "Correct", value: "good" },
  { label: "Facile", value: "easy" }
];

export function ReviewPanel({ card, activeIndex, totalCards, isLoading, onAnswer }) {
  if (isLoading) {
    return (
      <section className="review-panel" aria-label="Carte de revision">
        <p className="eyebrow">Carte active</p>
        <div className="card-face">
          <h2>Chargement</h2>
          <p>Preparation de la session</p>
        </div>
      </section>
    );
  }

  if (!card) {
    return (
      <section className="review-panel" aria-label="Carte de revision">
        <div className="review-header">
          <p className="eyebrow">Carte active</p>
          <span>0 / 0</span>
        </div>
        <div className="card-face">
          <h2>Aucune carte</h2>
          <p>Les flashcards seront chargees depuis le backend.</p>
        </div>
      </section>
    );
  }

  return (
    <section className="review-panel" aria-label="Carte de revision">
      <div className="review-header">
        <p className="eyebrow">Carte active</p>
        <span>
          {activeIndex + 1} / {totalCards}
        </span>
      </div>

      <div className="card-face">
        <h2>{card.korean}</h2>
        <p>{card.romanization || "romanisation a completer"}</p>
      </div>

      <div className="answer">
        <span>Traduction</span>
        <strong>{card.translation}</strong>
      </div>

      <p className="example">{card.exampleKorean || "Ajoute un exemple pour mieux retenir cette carte."}</p>

      <div className="rating-row" aria-label="Notation de la carte">
        {ratingOptions.map((option) => (
          <button key={option.value} type="button" onClick={() => onAnswer(option.value)}>
            {option.label}
          </button>
        ))}
      </div>
    </section>
  );
}
