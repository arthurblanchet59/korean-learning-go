import { useState } from "react";

import { createJournalEntry, deleteJournalEntry, previewJournalCorrection, updateJournalEntry } from "../../shared/api/studyApi.js";

export function JournalPanel({ entries, isMutating, runMutation, token }) {
  const [editingId, setEditingId] = useState(null);
  const [form, setForm] = useState({ title: "", text: "" });
  const [preview, setPreview] = useState(null);
  const [isChecking, setIsChecking] = useState(false);
  const [correctionError, setCorrectionError] = useState("");

  async function save(event) {
    event.preventDefault();
    const result = await runMutation(() => editingId
      ? updateJournalEntry(editingId, form, token)
      : createJournalEntry(form, token));
    if (result.ok) resetEditor();
  }

  async function correct() {
    if (!form.text.trim()) {
      setCorrectionError("Écris une phrase avant de demander une correction.");
      return;
    }
    setIsChecking(true);
    setCorrectionError("");
    const result = await previewJournalCorrection(form, token);
    if (result.ok) {
      setPreview(result.data);
    } else {
      setCorrectionError(result.error || "La correction automatique est indisponible.");
    }
    setIsChecking(false);
  }

  function edit(entry) {
    setEditingId(entry.id);
    setForm({ title: entry.title, text: entry.originalText });
    setPreview({
      correctedText: entry.correctedText,
      corrections: entry.corrections ?? [],
      sources: entry.sources ?? []
    });
  }

  function resetEditor() {
    setForm({ title: "", text: "" });
    setEditingId(null);
    setPreview(null);
    setCorrectionError("");
  }

  return (
    <fieldset aria-busy={isMutating} className="mutation-surface" disabled={isMutating}>
      <div className="journal-layout">
        <section className="management-section journal-editor">
          <p className="eyebrow">Expression écrite</p>
          <h2>{editingId ? "Modifier l'entrée" : "Écrire en coréen"}</h2>
          <form onSubmit={save}>
            <input maxLength="120" onChange={(event) => setForm({ ...form, title: event.target.value })} placeholder="Titre (facultatif)" value={form.title} />
            <textarea
              lang="ko"
              maxLength="10000"
              onChange={(event) => {
                setForm({ ...form, text: event.target.value });
                setPreview(null);
                setCorrectionError("");
              }}
              placeholder="오늘은 무엇을 했어요?"
              required
              rows="10"
              value={form.text}
            />
            <div className="button-row">
              <button className="secondary-button" disabled={isChecking} onClick={correct} type="button">{isChecking ? "Correction..." : "Vérifier le texte"}</button>
              <button className="primary-button" type="submit">Enregistrer</button>
              {editingId && <button className="text-button" onClick={resetEditor} type="button">Annuler</button>}
            </div>
          </form>
          {correctionError && <p className="form-error" role="alert">{correctionError}</p>}
          {preview && (
            <div className="correction-preview">
              <p className="eyebrow">Proposition corrigée</p>
              <p className="korean-text">{preview.correctedText}</p>
              {(preview.corrections ?? []).length === 0
                ? <span>Aucune correction automatique détectée.</span>
                : <ul>{preview.corrections.map((item, index) => <li key={`${item.original}-${index}`}><strong>{item.original} → {item.replacement}</strong><span>{item.reason}</span></li>)}</ul>}
              <CorrectionSources sources={preview.sources} />
            </div>
          )}
        </section>

        <section className="management-section journal-history">
          <div className="section-heading"><div><p className="eyebrow">Historique</p><h2>Mon journal</h2></div><strong>{entries.length}</strong></div>
          <div className="data-list">
            {entries.map((entry) => (
              <article className="journal-entry" key={entry.id}>
                <div><strong>{entry.title}</strong><time>{new Date(entry.createdAt).toLocaleDateString("fr-FR")}</time></div>
                <p className="korean-text">{entry.correctedText}</p>
                <small>{(entry.corrections ?? []).length} suggestion(s) · {(entry.sources ?? []).length} source(s)</small>
                <CorrectionSources compact sources={entry.sources} />
                <div className="button-row">
                  <button className="secondary-button" onClick={() => edit(entry)} type="button">Modifier</button>
                  <button className="danger-button" onClick={() => window.confirm(`Supprimer « ${entry.title} » ?`) && runMutation(() => deleteJournalEntry(entry.id, token))} type="button">Supprimer</button>
                </div>
              </article>
            ))}
          </div>
        </section>
      </div>
    </fieldset>
  );
}

function CorrectionSources({ compact = false, sources = [] }) {
  if (!sources?.length) return null;

  return (
    <div className={`correction-sources${compact ? " compact" : ""}`}>
      <strong>Leçons utilisées</strong>
      <div>
        {sources.map((source) => (
          <article key={source.id}>
            <span>{source.level} · {source.title}</span>
            {!compact && <p>{source.excerpt}</p>}
          </article>
        ))}
      </div>
    </div>
  );
}
