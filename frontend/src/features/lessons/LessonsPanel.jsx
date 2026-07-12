import { useEffect, useMemo, useState } from "react";

import { updateLessonProgress } from "../../shared/api/studyApi.js";

const sectionTitles = new Set(["OBJECTIF", "RÈGLE", "MÉTHODE", "EXEMPLES", "MODÈLE", "À RETENIR", "PRATIQUE", "CORRIGÉ"]);

export function LessonsPanel({ lessons, runMutation, token }) {
  const [activeId, setActiveId] = useState(lessons[0]?.id ?? "");
  const active = lessons.find((lesson) => lesson.id === activeId) ?? lessons[0];
  const [score, setScore] = useState(100);
  const [showCorrection, setShowCorrection] = useState(false);

  useEffect(() => {
    setScore(active?.progress?.score || 100);
    setShowCorrection(false);
  }, [active?.id, active?.progress?.score]);

  const sections = useMemo(() => parseLesson(active?.content ?? ""), [active?.content]);

  if (!active) return <section className="management-section empty-state"><h2>Aucune lecon</h2></section>;

  return (
    <section className="lesson-layout">
      <aside className="lesson-list">
        {lessons.map((lesson) => (
          <button className={lesson.id === active.id ? "active" : ""} key={lesson.id} onClick={() => setActiveId(lesson.id)} type="button">
            <span>{lesson.level}</span><strong>{lesson.title}</strong><small>{lesson.progress?.completed ? `Terminee · ${lesson.progress.score}%` : "A commencer"}</small>
          </button>
        ))}
      </aside>
      <article className="lesson-content">
        <p className="eyebrow">{active.level} · Lecon {active.order}</p>
        <h2>{active.title}</h2>
        <p className="lead">{active.description}</p>
        <div className="lesson-body">
          {sections.map((section, index) => {
            if (section.title === "CORRIGÉ" && !showCorrection) {
              return <button className="secondary-button correction-toggle" key={section.title} onClick={() => setShowCorrection(true)} type="button">Voir le corrige</button>;
            }
            return (
              <section className={section.title === "CORRIGÉ" ? "lesson-section correction" : "lesson-section"} key={`${section.title}-${index}`}>
                {section.title && <h3>{section.title}</h3>}
                <p>{section.body}</p>
              </section>
            );
          })}
        </div>
        <div className="lesson-actions">
          <label htmlFor={`score-${active.id}`}>Auto-evaluation
            <input id={`score-${active.id}`} max="100" min="0" onChange={(event) => setScore(Number(event.target.value))} type="number" value={score} />
          </label>
          <button className="primary-button" onClick={() => runMutation(() => updateLessonProgress(active.id, { completed: true, score }, token))} type="button">Marquer comme terminee</button>
        </div>
      </article>
    </section>
  );
}

function parseLesson(content) {
  return content.split(/\n\s*\n/).map((block) => {
    const lines = block.trim().split("\n");
    const candidate = lines[0]?.trim();
    if (sectionTitles.has(candidate)) {
      return { title: candidate, body: lines.slice(1).join("\n").trim() };
    }
    return { title: "", body: block.trim() };
  }).filter((section) => section.body);
}
