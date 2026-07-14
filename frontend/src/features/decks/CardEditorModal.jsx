import { useEffect, useRef } from "react";

export function CardEditorModal({ decks, form, isEditing, onChange, onClose, onSubmit, open }) {
  const firstField = useRef(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    if (!open) return undefined;
    firstField.current?.focus();
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const closeOnEscape = (event) => {
      if (event.key === "Escape") onCloseRef.current();
    };
    window.addEventListener("keydown", closeOnEscape);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  if (!open) return null;

  function update(field, value) {
    onChange((current) => ({ ...current, [field]: value }));
  }

  return (
    <div className="modal-backdrop" onMouseDown={(event) => event.target === event.currentTarget && onClose()} role="presentation">
      <section aria-labelledby="card-editor-title" aria-modal="true" className="card-editor-modal" role="dialog">
        <header className="modal-header">
          <div><p className="eyebrow">Bibliothèque</p><h2 id="card-editor-title">{isEditing ? "Modifier la carte" : "Ajouter une carte"}</h2></div>
          <button aria-label="Fermer" className="modal-close" onClick={onClose} title="Fermer" type="button">×</button>
        </header>

        <form className="card-editor-form" onSubmit={onSubmit}>
          <label>Deck<select onChange={(event) => update("deckId", event.target.value)} ref={firstField} required value={form.deckId}><option value="">Choisir un deck</option>{decks.map((deck) => <option key={deck.id} value={deck.id}>{deck.name}</option>)}</select></label>
          <label>Type<select onChange={(event) => update("kind", event.target.value)} value={form.kind}><option value="vocabulary">Vocabulaire</option><option value="phrase">Phrase</option><option value="hangul">Hangeul</option></select></label>
          <label>Coréen<input onChange={(event) => update("korean", event.target.value)} required value={form.korean} /></label>
          <label>Traduction<input onChange={(event) => update("translation", event.target.value)} required value={form.translation} /></label>
          <label>Romanisation<input onChange={(event) => update("romanization", event.target.value)} value={form.romanization} /></label>
          <label>Tags<input onChange={(event) => update("tags", event.target.value)} placeholder="salutation, débutant" value={form.tags} /></label>
          <label className="full-width">Exemple en coréen<textarea onChange={(event) => update("exampleKorean", event.target.value)} rows="2" value={form.exampleKorean} /></label>
          <label className="full-width">Traduction de l’exemple<textarea onChange={(event) => update("exampleTranslation", event.target.value)} rows="2" value={form.exampleTranslation} /></label>
          <footer className="modal-actions full-width"><button className="secondary-button" onClick={onClose} type="button">Annuler</button><button className="primary-button" type="submit">{isEditing ? "Enregistrer les modifications" : "Ajouter la carte"}</button></footer>
        </form>
      </section>
    </div>
  );
}
