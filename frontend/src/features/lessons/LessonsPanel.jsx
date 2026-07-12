import { useState } from "react";

import { updateLessonProgress } from "../../shared/api/studyApi.js";

export function LessonsPanel({ lessons, runMutation, token }) {
  const [activeId, setActiveId] = useState(lessons[0]?.id ?? "");
  const active = lessons.find((lesson) => lesson.id === activeId) ?? lessons[0];

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
        <div className="lesson-body">{active.content}</div>
        <div className="lesson-actions">
          <label>Score<input max="100" min="0" type="number" defaultValue={active.progress?.score ?? 100} id={`score-${active.id}`} /></label>
          <button className="primary-button" onClick={() => {
            const score = Number(document.getElementById(`score-${active.id}`).value);
            runMutation(() => updateLessonProgress(active.id, { completed: true, score }, token));
          }} type="button">Marquer comme terminee</button>
        </div>
      </article>
    </section>
  );
}
