import { useEffect, useState } from "react";

const ratingOptions = [
  { label: "Encore", value: "again" },
  { label: "Difficile", value: "hard" },
  { label: "Correct", value: "good" },
  { label: "Facile", value: "easy" }
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
        <h2>{isLoading ? "Chargement de la session" : "Revision terminee"}</h2>
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
    <section className="review-panel" aria-label="Carte de revision">
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
        {direction === "korean-to-french" && <p>{card.romanization}</p>}
      </div>

      {!revealed ? (
        <form className="answer-form" onSubmit={submitAnswer}>
          <label htmlFor="study-answer">Ta reponse</label>
          <div>
            <input autoComplete="off" id="study-answer" onChange={(event) => setAnswer(event.target.value)} required value={answer} />
            <button className="primary-button" type="submit">Verifier</button>
          </div>
          <button className="text-button" onClick={() => setRevealed(true)} type="button">Afficher sans repondre</button>
        </form>
      ) : (
        <>
          <div className={result?.correct ? "answer correct" : "answer"}>
            <span>{result ? (result.correct ? "Bonne reponse" : "A revoir") : "Reponse"}</span>
            <strong>{expected}</strong>
          </div>
          {card.exampleKorean && <p className="example">{card.exampleKorean}<br /><span>{card.exampleTranslation}</span></p>}
          <div className="rating-row" aria-label="Notation de la carte">
            {ratingOptions.map((option) => (
              <button key={option.value} onClick={() => onAnswer(option.value)} type="button">{option.label}</button>
            ))}
          </div>
        </>
      )}
    </section>
  );
}
