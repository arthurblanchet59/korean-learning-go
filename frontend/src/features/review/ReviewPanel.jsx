import { useEffect, useState } from "react";

const ratingOptions = [
  { label: "À revoir", description: "Je ne savais pas", value: "again" },
  { label: "Avec hésitation", description: "Réponse difficile", value: "hard" },
  { label: "Bien retenue", description: "Bonne réponse", value: "good" },
  { label: "Maîtrisée", description: "Réponse immédiate", value: "easy" }
];

export function ReviewPanel({ card, activeIndex, totalCards, isLoading, onAnswer, onCheck }) {
  const [direction, setDirection] = useState("korean-to-french");
  const [answer, setAnswer] = useState("");
  const [result, setResult] = useState(null);
  const [revealed, setRevealed] = useState(false);

  useEffect(() => {
    setAnswer("");
    setResult(null);
    setRevealed(false);
  }, [card?.id, direction]);

  if (isLoading || !card) {
    return (
      <section className="review-panel empty-state">
        <h2>{isLoading ? "Chargement de la session" : "Révision terminée"}</h2>
        <p>{isLoading ? "Lecture de ta progression..." : "Aucune carte n'est due pour le moment."}</p>
      </section>
    );
  }

  const prompt = direction === "korean-to-french" ? card.korean : card.translation;
  const expected = direction === "korean-to-french" ? card.translation : card.korean;

  async function submitAnswer(event) {
    event.preventDefault();
    const checked = await onCheck(card.id, answer, direction);
    if (checked) {
      setResult(checked);
      setRevealed(true);
    }
  }

  return (
    <section className="review-panel" aria-label="Carte de révision">
      <div className="review-header">
        <div>
          <p className="eyebrow">Carte active</p>
          <strong>{activeIndex + 1} / {totalCards}</strong>
        </div>
        <div className="segmented-control">
          <button className={direction === "korean-to-french" ? "active" : ""} onClick={() => setDirection("korean-to-french")} type="button">KO → FR</button>
          <button className={direction === "french-to-korean" ? "active" : ""} onClick={() => setDirection("french-to-korean")} type="button">FR → KO</button>
        </div>
      </div>

      <div className="card-face">
        <h2>{prompt}</h2>
      </div>

      {!revealed ? (
        <form className="answer-form" onSubmit={submitAnswer}>
          <label htmlFor="study-answer">Ta réponse</label>
          <div>
            <input autoComplete="off" id="study-answer" onChange={(event) => setAnswer(event.target.value)} required value={answer} />
            <button className="primary-button" type="submit">Vérifier</button>
          </div>
          <button className="text-button" onClick={() => setRevealed(true)} type="button">Afficher sans répondre</button>
        </form>
      ) : (
        <>
          <div className={result?.correct ? "answer correct" : "answer"}>
            <span>{result ? (result.correct ? "Bonne réponse" : "À revoir") : "Réponse"}</span>
            <strong>{expected}</strong>
            {card.romanization && <small>{card.romanization}</small>}
          </div>
          {card.exampleKorean && <p className="example">{card.exampleKorean}<br /><span>{card.exampleTranslation}</span></p>}
          <div className="rating-row" aria-label="Notation de la carte">
            {ratingOptions.map((option) => (
              <button key={option.value} onClick={() => onAnswer(option.value)} type="button"><strong>{option.label}</strong><small>{option.description}</small></button>
            ))}
          </div>
        </>
      )}
    </section>
  );
}
