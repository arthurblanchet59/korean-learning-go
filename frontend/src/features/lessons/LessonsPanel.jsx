import { useEffect, useMemo, useState } from "react";

import { completeLesson } from "../../shared/api/studyApi.js";

const sectionTitles = new Set(["OBJECTIF", "RÈGLE", "MÉTHODE", "EXEMPLES", "MODÈLE", "À RETENIR", "PRATIQUE", "CORRIGÉ"]);

export function LessonsPanel({ isMutating, lessons, runMutation, token }) {
  const [activeId, setActiveId] = useState(lessons[0]?.id ?? "");
  const active = lessons.find((lesson) => lesson.id === activeId) ?? lessons[0];
  const [showCorrection, setShowCorrection] = useState(false);

  useEffect(() => {
    setShowCorrection(false);
  }, [active?.id]);

  const sections = useMemo(() => parseLesson(active?.content ?? ""), [active?.content]);

  if (!active) return <section className="management-section empty-state"><h2>Aucune leçon</h2></section>;

  return (
    <fieldset aria-busy={isMutating} className="mutation-surface" disabled={isMutating}>
    <section className="lesson-layout">
      <aside className="lesson-list">
        {lessons.map((lesson) => (
          <button className={lesson.id === active.id ? "active" : ""} key={lesson.id} onClick={() => setActiveId(lesson.id)} type="button">
            <span>{lesson.level}</span><strong>{lesson.title}</strong><small>{lesson.progress?.completed ? "Terminée" : "À faire"}</small>
          </button>
        ))}
      </aside>
      <article className="lesson-content">
        <p className="eyebrow">{active.level} · Leçon {active.order}</p>
        <h2>{active.title}</h2>
        <p className="lead">{active.description}</p>
        <div className="lesson-body">
          {sections.map((section, index) => {
            if (section.title === "CORRIGÉ" && !showCorrection) {
              return <button className="secondary-button correction-toggle" key={section.title} onClick={() => setShowCorrection(true)} type="button">Voir le corrigé</button>;
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
          <span className={active.progress?.completed ? "lesson-status completed" : "lesson-status"}>{active.progress?.completed ? "Leçon terminée" : "Leçon à faire"}</span>
          <button className="primary-button" disabled={active.progress?.completed} onClick={() => runMutation(() => completeLesson(active.id, token))} type="button">{active.progress?.completed ? "Validée" : "Valider la leçon"}</button>
        </div>
      </article>
    </section>
    </fieldset>
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
